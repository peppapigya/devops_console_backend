package agent

import (
	"context"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/services/repair"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/bytedance/gopkg/util/logger"
)

const defaultMaxRounds = 10

// RunAgentLoop 核心 MCP Host Agent 循环（作为 repair.AgentRunFunc 传入 SessionService）
//
// 参数通过 repair 包类型传入，避免额外的接口层
func RunAgentLoop(
	sessionID string,
	logMessage, logHost, logService, logLevel string,
	creds repair.AgentSSHCreds,
	cfg repair.AgentLoopCfg,
	sessionMapper *mapper.RepairSessionMapper,
	msgMapper *mapper.SessionMessageMapper,
	actionMapper *mapper.RepairActionMapper,
	hub *repair.StreamHub,
) error {
	if err := sessionMapper.UpdateStatus(sessionID, "running"); err != nil {
		logger.Warnf("更新 session 状态失败: %v", err)
	}

	mcpClient := NewMCPServerClient(cfg.MCPServerURL, cfg.MCPToken)

	maxRounds := cfg.MaxRounds
	if maxRounds <= 0 {
		maxRounds = defaultMaxRounds
	}

	// ── 初始化 LLM Client ──
	llmCfg := LLMConfig{
		APIKey:  cfg.LLMAPIKey,
		BaseURL: cfg.LLMBaseURL,
		Model:   cfg.LLMModel,
	}
	llmClient := NewLLMClient(llmCfg)
	systemPrompt := BuildSystemPrompt(logHost)
	initialUserMsg := BuildInitialUserMessage(logMessage, logHost, logService, logLevel, creds.User, creds.Password, creds.Port)
	llmClient.SetSystemPrompt(systemPrompt)
	llmClient.AddUserMessage(initialUserMsg)

	// 持久化初始消息
	now := time.Now()
	_ = msgMapper.BatchCreate([]*model.SessionMessage{
		{SessionID: sessionID, Role: "system", Content: systemPrompt, CreatedAt: &now},
		{SessionID: sessionID, Role: "user", Content: initialUserMsg, CreatedAt: &now},
	})

	actionOrder := 0
	var lastActionSuccess bool = true
	var lastActionOutput string = ""

	for round := 0; round < maxRounds; round++ {
		// 推送 "AI 正在思考"
		hub.Publish(repair.SSEEvent{
			Type:      repair.EventThinking,
			SessionID: sessionID,
			Payload:   repair.ThinkingPayload{Content: fmt.Sprintf("第 %d 轮分析中，请稍候...", round+1)},
		})

		// ── 调用 LLM ──
		llmResp, err := llmClient.Call(context.Background())
		if err != nil {
			return fmt.Errorf("第 %d 轮 LLM 调用失败: %v", round+1, err)
		}

		// ══════════════════════════════════════════════════
		// Case 1：LLM 发出工具调用（execute_ssh）
		// ══════════════════════════════════════════════════
		if len(llmResp.ToolCalls) > 0 {
			for _, tc := range llmResp.ToolCalls {
				if tc.Name == "submit_diagnosis_report" {
					args := tc.Arguments
					fixed, _ := args["fixed"].(bool)

					// Guardrail：拦截 AI 的“修复成功”幻觉谎报
					if fixed && !lastActionSuccess {
						logger.Warnf("AI 正在说谎，试图提交 fixed:true，但上一条命令失败。进行拦截打回。")
						errMsg := fmt.Sprintf(`{"error":"【系统强力拦截】你声称已经修复成功（fixed: true），但你上一次执行的命令（验证命令）明明执行失败了！真实输出是：\n%s\n你是基于什么幻觉编造出验证通过的？严禁瞎编结果！请仔细研读上述真实的报错内容继续修复。如果实在无法修复，必须设置 fixed: false！"}`, lastActionOutput)
						llmClient.AddToolResult(tc.ID, errMsg)
						continue // 打回给大模型让它反思
					}

					// LLM 通过工具调用主动提交了最终结论
					conclusionStr, _ := json.Marshal(args)
					// 持久化 assistant 发出的结论信息
					msgNow := time.Now()
					_ = msgMapper.BatchCreate([]*model.SessionMessage{
						{SessionID: sessionID, Role: "assistant", Content: string(conclusionStr), CreatedAt: &msgNow},
					})

					// 提取关键字段
					args = tc.Arguments
					conclusion, _ := args["conclusion"].(string)
					rootCause, _ := args["root_cause"].(string)
					severity, _ := args["severity"].(string)
					if severity == "" {
						severity = "medium"
					}

					// 更新 session 分析字段
					_ = sessionMapper.UpdateFields(sessionID, map[string]interface{}{
						"analysis":   conclusion,
						"root_cause": rootCause,
						"severity":   severity,
						"confidence": 0.95,
					})

					// 推送计划摘要（前端展示最终报告）
					hub.Publish(repair.SSEEvent{
						Type:      repair.EventPlan,
						SessionID: sessionID,
						Payload: repair.PlanPayload{
							Analysis:   conclusion,
							RootCause:  rootCause,
							Severity:   severity,
							Confidence: 0.95,
						},
					})

					// 正常完成
					actions, _ := actionMapper.GetBySessionID(sessionID)
					completed := 0
					for _, a := range actions {
						if a.Status == "success" {
							completed++
						}
					}
					finAt := time.Now()
					_ = sessionMapper.UpdateFields(sessionID, map[string]interface{}{
						"status":      "success",
						"finished_at": &finAt,
					})
					hub.PublishDone(sessionID, repair.DonePayload{
						Status:           "success",
						CompletedActions: completed,
						TotalActions:     len(actions),
					})
					return nil // 退出 Agent 循环
				}

				if tc.Name != "execute_ssh" {
					llmClient.AddToolResult(tc.ID, fmt.Sprintf(`{"error":"未知工具: %s"}`, tc.Name))
					continue
				}

				// 提取 SSH 参数
				toolReq, riskLevel, description, thought, extractErr := extractSSHParams(tc.Arguments, logHost, creds)
				if extractErr != nil {
					llmClient.AddToolResult(tc.ID, fmt.Sprintf(`{"error":"%s"}`, extractErr.Error()))
					continue
				}

				actionOrder++
				createdAt := time.Now()

				// 持久化修复动作
				action := &model.RepairAction{
					SessionID:   sessionID,
					ActionOrder: actionOrder,
					Description: description,
					Thought:     thought,
					Command:     toolReq.Command,
					Cwd:         toolReq.Cwd,
					Target:      toolReq.Host,
					Timeout:     toolReq.Timeout,
					RiskLevel:   riskLevel,
					Status:      "pending",
					CreatedAt:   &createdAt,
				}
				if createErr := actionMapper.BatchCreate([]*model.RepairAction{action}); createErr != nil {
					logger.Warnf("持久化 action 失败: %v", createErr)
				}
				_ = sessionMapper.UpdateFields(sessionID, map[string]interface{}{
					"total_actions": actionOrder,
				})

				// ── 高风险动作：等待用户确认 ──
				if riskLevel == "high" {
					_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{"status": "waiting_confirm"})
					_ = sessionMapper.UpdateStatus(sessionID, "waiting_confirm")

					hub.Publish(repair.SSEEvent{
						Type:      repair.EventWaitConfirm,
						SessionID: sessionID,
						Payload: repair.WaitConfirmPayload{
							ActionID:    action.ID,
							ActionOrder: actionOrder,
							Description: description,
							Command:     toolReq.Command,
							RiskReason:  "该命令风险等级为 high，需要您确认后才会执行",
						},
					})

					// 注册等待并阻塞
					confirmCh := repair.RegisterConfirmWait(sessionID, action.ID)
					approved := false
					select {
					case approved = <-confirmCh:
					case <-time.After(10 * time.Minute):
					}

					if !approved {
						_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{"status": "skipped"})
						_ = sessionMapper.UpdateStatus(sessionID, "running")
						llmClient.AddToolResult(tc.ID, `{"success":false,"output":"","error":"用户拒绝执行该命令，请考虑其他修复方案"}`)
						continue
					}
					_ = sessionMapper.UpdateStatus(sessionID, "running")
				}

				// ── 执行工具 ──
				_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{"status": "running"})
				hub.Publish(repair.SSEEvent{
					Type:      repair.EventActionStart,
					SessionID: sessionID,
					Payload: repair.ActionStartPayload{
						ActionID:    action.ID,
						ActionOrder: actionOrder,
						ToolName:    tc.Name,
						Thought:     thought,
						Description: description,
						Command:     toolReq.Command,
						Target:      toolReq.Host,
					},
				})

				t0 := time.Now()
				toolResult, toolErr := mcpClient.ExecuteSSH(toolReq)
				finishedAt := time.Now()

				if toolErr != nil {
					_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{
						"status":      "failed",
						"error_msg":   toolErr.Error(),
						"exit_code":   -1,
						"duration_ms": finishedAt.Sub(t0).Milliseconds(),
						"executed_at": &finishedAt,
					})
					hub.Publish(repair.SSEEvent{
						Type:      repair.EventActionResult,
						SessionID: sessionID,
						Payload:   repair.ActionResultPayload{ActionID: action.ID, Status: "failed", ErrorMsg: toolErr.Error(), ExitCode: -1},
					})
					errJSON, _ := json.Marshal(map[string]interface{}{"success": false, "error": toolErr.Error(), "exit_code": -1})
					llmClient.AddToolResult(tc.ID, string(errJSON))
					continue
				}

				// 工具执行完成
				status := "success"
				if !toolResult.Success {
					status = "failed"
				}
				lastActionSuccess = toolResult.Success
				lastActionOutput = toolResult.Output + "\n" + toolResult.Stderr
				durationMs := int(toolResult.DurationMs)
				_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{
					"status":      status,
					"output":      toolResult.Output + "\n" + toolResult.Stderr,
					"exit_code":   toolResult.ExitCode,
					"duration_ms": durationMs,
					"executed_at": &finishedAt,
				})
				_ = sessionMapper.UpdateFields(sessionID, map[string]interface{}{"completed_actions": actionOrder})

				hub.Publish(repair.SSEEvent{
					Type:      repair.EventActionResult,
					SessionID: sessionID,
					Payload: repair.ActionResultPayload{
						ActionID:   action.ID,
						Status:     status,
						Output:     toolResult.Output,
						ErrorMsg:   toolResult.Stderr,
						ExitCode:   toolResult.ExitCode,
						DurationMs: durationMs,
					},
				})

				// 将执行结果注入 LLM 对话历史
				resultJSON, _ := json.Marshal(map[string]interface{}{
					"success":     toolResult.Success,
					"output":      toolResult.Output,
					"stderr":      toolResult.Stderr,
					"exit_code":   toolResult.ExitCode,
					"duration_ms": toolResult.DurationMs,
				})
				llmClient.AddToolResult(tc.ID, string(resultJSON))

				// 持久化本轮 tool result 消息
				msgNow := time.Now()
				_ = msgMapper.BatchCreate([]*model.SessionMessage{
					{
						SessionID: sessionID,
						Role:      "tool",
						Content:   string(resultJSON),
						CreatedAt: &msgNow,
					},
				})
			}
			// 继续下一轮（让 LLM 分析工具结果）
			continue
		}

		// ══════════════════════════════════════════════════
		// Case 2：LLM 输出纯文本（逃避工具调用）
		// ══════════════════════════════════════════════════
		if llmResp.Content != "" {
			// 持久化 assistant 纯文本回复
			msgNow := time.Now()
			_ = msgMapper.BatchCreate([]*model.SessionMessage{
				{SessionID: sessionID, Role: "assistant", Content: llmResp.Content, CreatedAt: &msgNow},
			})

			// 强制打回，要求使用工具
			logger.Warnf("AI 逃避调用工具，输出纯文本，拦截指令：\n%s", llmResp.Content)
			warningMsg := "【系统强力拦截】你违反了协议配置：不允许只回复纯文本而不调用底层工具！如果你想执行命令，请调用 `execute_ssh`；如果你认为排查已经结束，请务必调用 `submit_diagnosis_report` 工具正式提交结论！如果由于命令持续报错无法修复，请主动调用 `submit_diagnosis_report` 并传入 fixed:false！"
			llmClient.AddUserMessage(warningMsg)

			msgNow2 := time.Now()
			_ = msgMapper.BatchCreate([]*model.SessionMessage{
				{SessionID: sessionID, Role: "user", Content: warningMsg, CreatedAt: &msgNow2},
			})
			continue
		}

		// LLM 既无工具调用也无合法内容，降级处理
		logger.Warnf("[AgentLoop] 第 %d 轮 LLM 回复无效，内容=%s", round+1, llmResp.Content)
		if llmResp.Content != "" {
			llmClient.AddUserMessage("请继续分析，或输出最终诊断报告。")
		}
	}

	return fmt.Errorf("Agent 已达最大诊断轮数（%d轮），诊断未完成，请手动检查", maxRounds)
}

// ────────────────────────────────────────────────────────────────────
// 内部辅助函数
// ────────────────────────────────────────────────────────────────────

func extractSSHParams(args map[string]interface{}, defaultHost string, creds repair.AgentSSHCreds) (ToolCallRequest, string, string, string, error) {
	host, _ := args["host"].(string)
	if host == "" {
		host = defaultHost
	}
	user, _ := args["user"].(string)
	if user == "" {
		user = creds.User
	}
	password, _ := args["password"].(string)
	if password == "" {
		password = creds.Password
	}
	command, _ := args["command"].(string)
	if command == "" {
		return ToolCallRequest{}, "", "", "", fmt.Errorf("command 参数不能为空")
	}

	cwd, _ := args["cwd"].(string)
	riskLevel, _ := args["risk_level"].(string)
	if riskLevel == "" {
		riskLevel = "low"
	}
	description, _ := args["description"].(string)

	port := creds.Port
	if portRaw, ok := args["port"]; ok {
		switch v := portRaw.(type) {
		case float64:
			port = int(v)
		case string:
			if p, err := strconv.Atoi(v); err == nil {
				port = p
			}
		}
	}

	timeout := 30
	if timeoutRaw, ok := args["timeout"]; ok {
		if v, ok := timeoutRaw.(float64); ok {
			timeout = int(v)
		}
	}

	thought, _ := args["thought"].(string)

	return ToolCallRequest{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Command:  command,
		Cwd:      cwd,
		Timeout:  timeout,
	}, riskLevel, description, thought, nil
}

type finalConclusion struct {
	Conclusion     string `json:"conclusion"`
	RootCause      string `json:"root_cause"`
	Severity       string `json:"severity"`
	Fixed          bool   `json:"fixed"`
	FixSummary     string `json:"fix_summary"`
	Recommendation string `json:"recommendation"`
}

func parseFinalConclusion(content string) finalConclusion {
	var c finalConclusion
	if err := json.Unmarshal([]byte(content), &c); err == nil {
		return c
	}
	return finalConclusion{
		Conclusion: content,
		RootCause:  "见诊断报告",
		Severity:   "medium",
	}
}
