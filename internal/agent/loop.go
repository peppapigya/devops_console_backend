package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	monitorCtrl "devops-console-backend/internal/controllers/monitor"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/services/repair"

	"github.com/bytedance/gopkg/util/logger"
)

const defaultMaxRounds = 12

type reportConclusion struct {
	Summary        string   `json:"summary"`
	RootCause      string   `json:"root_cause"`
	Severity       string   `json:"severity"`
	Fixed          bool     `json:"fixed"`
	Symptoms       []string `json:"symptoms"`
	Evidence       []string `json:"evidence"`
	FixSummary     string   `json:"fix_summary"`
	Recommendation string   `json:"recommendation"`
	NextSteps      []string `json:"next_steps"`
}

type toolExecution struct {
	name        string
	request     interface{}
	commandText string
	target      string
	riskLevel   string
	riskReason  string
	description string
	thought     string
}

type validationBlocker struct {
	rawOutput string
	line      string
	summary   string
}

func RunAgentLoop(sessionID string, logMessage, logHost, logService, logLevel string, creds repair.AgentSSHCreds, cfg repair.AgentLoopCfg, sessionMapper *mapper.RepairSessionMapper, msgMapper *mapper.SessionMessageMapper, actionMapper *mapper.RepairActionMapper, hub *repair.StreamHub) error {
	if err := sessionMapper.UpdateStatus(sessionID, "analyzing"); err != nil {
		logger.Warnf("update session status failed: %v", err)
	}

	mcpClient := NewMCPServerClient(cfg.MCPServerURL, cfg.MCPToken)
	maxRounds := cfg.MaxRounds
	if maxRounds <= 0 {
		maxRounds = defaultMaxRounds
	}

	llmClient := NewLLMClient(LLMConfig{APIKey: cfg.LLMAPIKey, BaseURL: cfg.LLMBaseURL, Model: cfg.LLMModel})
	systemPrompt := BuildSystemPrompt(logHost)
	initialUserMsg := BuildInitialUserMessage(logMessage, logHost, logService, logLevel, creds.User, creds.Password, creds.Port)
	resourceGuide := "在开始现场取证前，优先调用 read_service_resource 读取与当前服务同名的知识资源；若资源不存在，再尝试 read_knowledge_base，最后才结合现场证据做保守推理。"
	llmClient.SetSystemPrompt(systemPrompt)
	llmClient.AddUserMessage(initialUserMsg)
	llmClient.AddUserMessage(resourceGuide)

	now := time.Now()
	_ = msgMapper.BatchCreate([]*model.SessionMessage{
		{SessionID: sessionID, Role: "system", Content: systemPrompt, CreatedAt: &now},
		{SessionID: sessionID, Role: "user", Content: initialUserMsg, CreatedAt: &now},
		{SessionID: sessionID, Role: "user", Content: resourceGuide, CreatedAt: &now},
	})

	actionOrder := 0
	lastActionSuccess := true
	lastActionOutput := ""
	lastValidationSucceeded := false
	commandAttempts := map[string]int{}
	serviceResourceHit := false
	postWriteReadRequired := false
	lastWrittenFile := ""
	currentValidationBlocker := validationBlocker{}

	for round := 0; round < maxRounds; round++ {
		hub.Publish(repair.SSEEvent{
			Type:      repair.EventThinking,
			SessionID: sessionID,
			Payload:   repair.ThinkingPayload{Content: fmt.Sprintf("第 %d 轮：正在收集证据并评估下一步...", round+1)},
		})

		llmResp, err := llmClient.Call(context.Background())
		if err != nil {
			return fmt.Errorf("round %d llm call failed: %v", round+1, err)
		}

		if len(llmResp.ToolCalls) > 0 {
			for _, tc := range llmResp.ToolCalls {
				if tc.Name == "submit_diagnosis_report" {
					if err := handleDiagnosisReport(tc, sessionID, actionMapper, sessionMapper, msgMapper, hub, lastActionSuccess, lastActionOutput, llmClient); err != nil {
						return err
					}
					return nil
				}

				exec, err := buildToolExecution(tc, logHost, creds)
				if err != nil {
					llmClient.AddToolResult(tc.ID, fmt.Sprintf(`{"error":%q}`, err.Error()))
					continue
				}

				if exec.name == "read_knowledge_base" {
					if req, ok := exec.request.(KBRequest); ok && strings.EqualFold(strings.TrimSpace(req.Topic), strings.TrimSpace(logService)) && serviceResourceHit {
						blockMsg := fmt.Sprintf("当前服务 %s 已成功命中 service resource，不要再重复读取同主题 knowledge base；请直接依据资源内容继续取证和修复。", logService)
						llmClient.AddToolResult(tc.ID, fmt.Sprintf(`{"success":false,"error":%q}`, blockMsg))
						llmClient.AddUserMessage(blockMsg)
						continue
					}
				}

				if postWriteReadRequired && exec.name != "inspect_file_snippet" {
					blockMsg := fmt.Sprintf("刚刚修改过文件 %s，下一步必须先重新读取同一文件确认改动，再继续其他动作。", lastWrittenFile)
					llmClient.AddToolResult(tc.ID, fmt.Sprintf(`{"success":false,"error":%q}`, blockMsg))
					llmClient.AddUserMessage(blockMsg)
					continue
				}

				commandKey := strings.TrimSpace(exec.name + "::" + exec.commandText)
				if commandKey != "::" && commandAttempts[commandKey] >= 2 {
					msg := "检测到重复执行同一命令。请基于新假设选择不同工具，或直接调用 submit_diagnosis_report 提交结论。"
					llmClient.AddToolResult(tc.ID, fmt.Sprintf(`{"success":false,"error":%q}`, msg))
					llmClient.AddUserMessage("不要重复同一命令。请改用不同证据路径，或立即调用 submit_diagnosis_report 结束流程。")
					continue
				}
				if commandKey != "::" {
					commandAttempts[commandKey]++
				}

				actionOrder++
				createdAt := time.Now()
				action := &model.RepairAction{
					SessionID:   sessionID,
					ActionOrder: actionOrder,
					ToolName:    exec.name,
					Description: exec.description,
					Thought:     exec.thought,
					Command:     exec.commandText,
					Target:      exec.target,
					Timeout:     extractTimeout(exec.request),
					RiskLevel:   exec.riskLevel,
					RiskReason:  exec.riskReason,
					Status:      "pending",
					CreatedAt:   &createdAt,
				}
				if err := actionMapper.BatchCreate([]*model.RepairAction{action}); err != nil {
					logger.Warnf("persist action failed: %v", err)
				}
				monitorCtrl.RepairActionsTotal.WithLabelValues(exec.name, "created", exec.riskLevel).Inc()
				publishSessionProgress(sessionMapper, hub, sessionID, "analyzing", actionOrder, actionOrder-1)

				if exec.name == "restart_service" && !lastValidationSucceeded {
					blockMsg := "禁止执行服务重启：最近一次验证未成功，必须先完成成功验证后才能重启服务。"
					_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{
						"status":      "failed",
						"error_msg":   blockMsg,
						"exit_code":   -1,
						"duration_ms": 0,
					})
					monitorCtrl.RepairActionsTotal.WithLabelValues(exec.name, "failed", exec.riskLevel).Inc()
					hub.Publish(repair.SSEEvent{
						Type:      repair.EventActionResult,
						SessionID: sessionID,
						Payload: repair.ActionResultPayload{
							ActionID: action.ID,
							Status:   "failed",
							ErrorMsg: blockMsg,
							ExitCode: -1,
						},
					})
					llmClient.AddToolResult(tc.ID, fmt.Sprintf(`{"success":false,"error":%q,"exit_code":-1}`, blockMsg))
					llmClient.AddUserMessage("不要在配置验证失败后继续重启服务。请重新读取目标文件，确认实际内容与待修改片段完全一致，再修复并重新验证。")
					continue
				}

				if exec.riskLevel == "high" {
					_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{"status": "waiting_confirm"})
					publishSessionProgress(sessionMapper, hub, sessionID, "waiting_approval", actionOrder, actionOrder-1)
					hub.Publish(repair.SSEEvent{
						Type:      repair.EventWaitConfirm,
						SessionID: sessionID,
						Payload: repair.WaitConfirmPayload{
							ActionID:    action.ID,
							ActionOrder: actionOrder,
							Description: exec.description,
							Command:     exec.commandText,
							RiskReason:  exec.riskReason,
						},
					})
					confirmCh := repair.RegisterConfirmWait(sessionID, action.ID)
					approved := false
					select {
					case approved = <-confirmCh:
					case <-time.After(10 * time.Minute):
					}
					if !approved {
						_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{"status": "skipped"})
						monitorCtrl.RepairActionsTotal.WithLabelValues(exec.name, "skipped", exec.riskLevel).Inc()
						publishSessionProgress(sessionMapper, hub, sessionID, "analyzing", actionOrder, actionOrder-1)
						llmClient.AddToolResult(tc.ID, `{"success":false,"error":"user rejected the high-risk action"}`)
						continue
					}
					publishSessionProgress(sessionMapper, hub, sessionID, "executing", actionOrder, actionOrder-1)
				}

				_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{"status": "running"})
				hub.Publish(repair.SSEEvent{
					Type:      repair.EventActionStart,
					SessionID: sessionID,
					Payload: repair.ActionStartPayload{
						ActionID:    action.ID,
						ActionOrder: actionOrder,
						ToolName:    exec.name,
						Thought:     exec.thought,
						Description: exec.description,
						Command:     exec.commandText,
						Target:      exec.target,
					},
				})

				toolResult, toolErr := executeStructuredTool(mcpClient, exec)
				finishedAt := time.Now()
				if toolErr != nil {
					if exec.name == "validate_nginx_config" {
						lastValidationSucceeded = false
					}
					_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{
						"status":      "failed",
						"error_msg":   toolErr.Error(),
						"exit_code":   -1,
						"duration_ms": 0,
						"executed_at": &finishedAt,
					})
					monitorCtrl.RepairActionsTotal.WithLabelValues(exec.name, "failed", exec.riskLevel).Inc()
					hub.Publish(repair.SSEEvent{
						Type:      repair.EventActionResult,
						SessionID: sessionID,
						Payload: repair.ActionResultPayload{
							ActionID: action.ID,
							Status:   "failed",
							ErrorMsg: toolErr.Error(),
							ExitCode: -1,
						},
					})
					llmClient.AddToolResult(tc.ID, fmt.Sprintf(`{"success":false,"error":%q,"exit_code":-1}`, toolErr.Error()))
					if exec.name == "replace_file_content" {
						llmClient.AddUserMessage("文件修改工具执行失败，文件大概率尚未变化。请先重新读取文件当前内容，再基于精确原文重新替换，或改用其他可证明修改成功的方式。")
					}
					continue
				}

				status := "success"
				if !toolResult.Success {
					status = "failed"
				}
				if exec.name == "validate_nginx_config" {
					lastValidationSucceeded = toolResult.Success
				}
				if exec.name == "replace_file_content" {
					lastValidationSucceeded = false
					postWriteReadRequired = toolResult.Success
					if req, ok := exec.request.(FileReplaceRequest); ok {
						lastWrittenFile = req.FilePath
					}
				}
				if exec.name == "inspect_file_snippet" {
					if req, ok := exec.request.(FileSnippetRequest); ok && postWriteReadRequired && sameFile(req.FilePath, lastWrittenFile) {
						postWriteReadRequired = false
					}
				}
				if exec.name == "read_service_resource" && toolResult.Success {
					serviceResourceHit = true
				}

				lastActionSuccess = toolResult.Success
				rawActionOutput := strings.TrimSpace(toolResult.Output + "\n" + toolResult.Stderr)
				lastActionOutput = rawActionOutput
				if exec.name == "validate_nginx_config" {
					if toolResult.Success {
						currentValidationBlocker = validationBlocker{}
					} else {
						currentValidationBlocker = parseValidationBlocker(rawActionOutput)
					}
				}
				if len(lastActionOutput) > 1200 {
					lastActionOutput = lastActionOutput[:1200] + "..."
				}

				_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{
					"status":      status,
					"output":      rawActionOutput,
					"exit_code":   toolResult.ExitCode,
					"duration_ms": int(toolResult.DurationMs),
					"executed_at": &finishedAt,
				})
				monitorCtrl.RepairActionsTotal.WithLabelValues(exec.name, status, exec.riskLevel).Inc()
				monitorCtrl.RepairActionDuration.WithLabelValues(exec.name, status).Observe(float64(toolResult.DurationMs) / 1000.0)
				publishSessionProgress(sessionMapper, hub, sessionID, "executing", actionOrder, countCompletedActions(actionMapper, sessionID))
				hub.Publish(repair.SSEEvent{
					Type:      repair.EventActionResult,
					SessionID: sessionID,
					Payload: repair.ActionResultPayload{
						ActionID:   action.ID,
						Status:     status,
						Output:     toolResult.Output,
						ErrorMsg:   toolResult.Stderr,
						ExitCode:   toolResult.ExitCode,
						DurationMs: int(toolResult.DurationMs),
					},
				})

				compacted := CompactToolResult(exec.name, exec.commandText, toolResult)
				resultJSON, _ := json.Marshal(compacted)
				llmClient.AddToolResult(tc.ID, string(resultJSON))

				if exec.name == "read_service_resource" && toolResult.Success {
					llmClient.AddUserMessage("你已命中当前服务的专属知识资源。下一步请优先遵循 resource 中的服务专属排障规则继续取证和修复，不要自行脑补额外的服务语法规则。")
				}
				if exec.name == "read_service_resource" && !toolResult.Success {
					llmClient.AddUserMessage("当前服务没有命中的知识资源。请继续现场取证；如有相关通用主题，也可以改用 read_knowledge_base 读取知识库，再结合命令输出完成诊断。")
				}
				if exec.name == "inspect_file_snippet" {
					llmClient.AddUserMessage("如果下一步需要调用 replace_file_content，search 必须直接使用刚刚回读结果里的完整原始文本片段，不要凭推测改写目标行，也不要自行变更其他行。")
				}
				if exec.name == "validate_nginx_config" && toolResult.Success {
					llmClient.AddUserMessage("nginx 配置验证已成功。下一步必须调用 submit_diagnosis_report 提交最终结论，不要继续重复取证。")
				} else if exec.name == "validate_nginx_config" && !toolResult.Success {
					llmClient.AddUserMessage("配置验证失败，说明修复尚未生效。不要继续重启服务，也不要声称已修复成功；请重新读取文件内容，确认实际文件与错误行，再调整修复方案。")
					if currentValidationBlocker.summary != "" {
						llmClient.AddUserMessage(fmt.Sprintf("最新验证错误已经变化，当前阻塞问题是：%s。请先围绕这个新错误继续取证和修复，不要继续沿用旧报错的修复思路。", currentValidationBlocker.summary))
					}
				}
				if exec.name == "replace_file_content" && !toolResult.Success {
					llmClient.AddUserMessage("replace_file_content 返回失败时，文件当前内容大概率没有按预期改动。尤其当错误是 search text not found 时，必须重新读取文件，并使用精确匹配的原文后再尝试修改。")
				}
				if exec.name == "replace_file_content" && toolResult.Success {
					llmClient.AddUserMessage("文件修改已执行。下一步必须先重新读取同一文件确认实际改动结果，再执行验证命令；不要直接提交结论，也不要直接重启服务。")
				}
				if exec.name == "inspect_file_snippet" && currentValidationBlocker.summary != "" {
					llmClient.AddUserMessage(fmt.Sprintf("当前仍有未解决的最新验证错误：%s。请基于刚刚回读的原文，优先修复这个中间阻塞问题，再继续处理最终目标。", currentValidationBlocker.summary))
				}

				msgNow := time.Now()
				_ = msgMapper.BatchCreate([]*model.SessionMessage{{SessionID: sessionID, Role: "tool", Content: string(resultJSON), CreatedAt: &msgNow}})
			}

			if round >= maxRounds-2 {
				llmClient.AddUserMessage("你即将达到最大轮次。请立即收敛，并调用 submit_diagnosis_report 输出最终报告。")
			}
			continue
		}

		if llmResp.Content != "" {
			msgNow := time.Now()
			_ = msgMapper.BatchCreate([]*model.SessionMessage{{SessionID: sessionID, Role: "assistant", Content: llmResp.Content, CreatedAt: &msgNow}})
			llmClient.AddUserMessage("不要只输出纯文本。请调用工具收集证据，或调用 submit_diagnosis_report 结束流程，并确保推理描述使用中文。")
			continue
		}
	}

	return fmt.Errorf("agent reached max diagnosis rounds (%d) without completion", maxRounds)
}

func handleDiagnosisReport(tc ToolCall, sessionID string, actionMapper *mapper.RepairActionMapper, sessionMapper *mapper.RepairSessionMapper, msgMapper *mapper.SessionMessageMapper, hub *repair.StreamHub, lastActionSuccess bool, lastActionOutput string, llmClient *LLMClient) error {
	args := tc.Arguments
	fixed, _ := args["fixed"].(bool)
	if fixed && !lastActionSuccess {
		msg := fmt.Sprintf("最近一次验证或执行失败，不能提交 fixed=true。最近输出：%s", lastActionOutput)
		llmClient.AddToolResult(tc.ID, fmt.Sprintf(`{"error":%q}`, msg))
		return nil
	}

	payloadBytes, _ := json.Marshal(args)
	report := parseReportConclusion(payloadBytes)
	analysisText := buildAnalysisSummary(report)
	severity := report.Severity
	if severity == "" {
		severity = "medium"
	}
	confidence := 0.9
	if !report.Fixed {
		confidence = 0.72
	}

	msgNow := time.Now()
	_ = msgMapper.BatchCreate([]*model.SessionMessage{{SessionID: sessionID, Role: "assistant", Content: string(payloadBytes), CreatedAt: &msgNow}})
	_ = sessionMapper.UpdateFields(sessionID, map[string]interface{}{"analysis": analysisText, "root_cause": report.RootCause, "severity": severity, "confidence": confidence})

	hub.Publish(repair.SSEEvent{Type: repair.EventPlan, SessionID: sessionID, Payload: repair.PlanPayload{Analysis: analysisText, RootCause: report.RootCause, Severity: severity, Confidence: confidence}})
	completed := countCompletedActions(actionMapper, sessionID)
	actions, _ := actionMapper.GetBySessionID(sessionID)
	finAt := time.Now()
	finalStatus := "completed"
	if !report.Fixed {
		finalStatus = "failed"
	}
	_ = sessionMapper.UpdateFields(sessionID, map[string]interface{}{"status": finalStatus, "finished_at": &finAt, "completed_actions": completed, "total_actions": len(actions)})
	monitorCtrl.RepairSessionsTotal.WithLabelValues(finalStatus).Inc()
	hub.PublishDone(sessionID, repair.DonePayload{Status: finalStatus, CompletedActions: completed, TotalActions: len(actions)})
	return nil
}

func buildToolExecution(tc ToolCall, defaultHost string, creds repair.AgentSSHCreds) (*toolExecution, error) {
	switch tc.Name {
	case "read_service_resource":
		service, err := stringArg(tc.Arguments, "service")
		if err != nil {
			return nil, err
		}
		return &toolExecution{
			name:        tc.Name,
			request:     ServiceResourceRequest{Service: service},
			commandText: fmt.Sprintf("read service resource: %s", service),
			riskLevel:   strings.ToLower(optionalStringArg(tc.Arguments, "risk_level")),
			riskReason:  optionalStringArg(tc.Arguments, "risk_reason"),
			description: optionalStringArg(tc.Arguments, "description"),
			thought:     optionalStringArg(tc.Arguments, "thought"),
		}, nil
	case "read_knowledge_base":
		topic, err := stringArg(tc.Arguments, "topic")
		if err != nil {
			return nil, err
		}
		return &toolExecution{
			name:        tc.Name,
			request:     KBRequest{Topic: topic},
			commandText: fmt.Sprintf("read knowledge base: %s", topic),
			riskLevel:   strings.ToLower(optionalStringArg(tc.Arguments, "risk_level")),
			riskReason:  optionalStringArg(tc.Arguments, "risk_reason"),
			description: optionalStringArg(tc.Arguments, "description"),
			thought:     optionalStringArg(tc.Arguments, "thought"),
		}, nil
	case "write_knowledge_base":
		topic, err := stringArg(tc.Arguments, "topic")
		if err != nil {
			return nil, err
		}
		content, err := stringArg(tc.Arguments, "content")
		if err != nil {
			return nil, err
		}
		return &toolExecution{
			name:        tc.Name,
			request:     KBWriteRequest{Topic: topic, Content: content},
			commandText: fmt.Sprintf("write knowledge base: %s", topic),
			riskLevel:   strings.ToLower(optionalStringArg(tc.Arguments, "risk_level")),
			riskReason:  optionalStringArg(tc.Arguments, "risk_reason"),
			description: optionalStringArg(tc.Arguments, "description"),
			thought:     optionalStringArg(tc.Arguments, "thought"),
		}, nil
	}

	meta, err := extractToolMeta(tc.Arguments, defaultHost, creds)
	if err != nil {
		return nil, err
	}
	te := &toolExecution{name: tc.Name, target: meta.Host, riskLevel: meta.RiskLevel, riskReason: meta.RiskReason, description: meta.Description, thought: meta.Thought}

	switch tc.Name {
	case "inspect_service_status":
		service, err := stringArg(tc.Arguments, "service")
		if err != nil {
			return nil, err
		}
		te.request = ServiceStatusRequest{StructuredToolBase: meta.StructuredToolBase, Service: service}
		te.commandText = fmt.Sprintf("inspect service status: %s", service)
	case "inspect_service_logs":
		service, err := stringArg(tc.Arguments, "service")
		if err != nil {
			return nil, err
		}
		te.request = ServiceLogsRequest{StructuredToolBase: meta.StructuredToolBase, Service: service, Lines: intArg(tc.Arguments, "lines", 80)}
		te.commandText = fmt.Sprintf("inspect recent service logs: %s", service)
	case "inspect_file_snippet":
		filePath, err := stringArg(tc.Arguments, "file_path")
		if err != nil {
			return nil, err
		}
		lineStart := intArg(tc.Arguments, "line_start", 1)
		lineEnd := intArg(tc.Arguments, "line_end", lineStart)
		te.request = FileSnippetRequest{StructuredToolBase: meta.StructuredToolBase, FilePath: filePath, LineStart: lineStart, LineEnd: lineEnd}
		te.commandText = fmt.Sprintf("inspect file snippet: %s (%d-%d)", filePath, lineStart, lineEnd)
	case "validate_nginx_config":
		te.request = NginxConfigValidateRequest{StructuredToolBase: meta.StructuredToolBase, ConfigPath: optionalStringArg(tc.Arguments, "config_path")}
		te.commandText = "validate nginx config"
	case "replace_file_content":
		filePath, err := stringArg(tc.Arguments, "file_path")
		if err != nil {
			return nil, err
		}
		search, err := stringArg(tc.Arguments, "search")
		if err != nil {
			return nil, err
		}
		replace := optionalStringArg(tc.Arguments, "replace")
		createBackup := boolArg(tc.Arguments, "create_backup", true)
		te.request = FileReplaceRequest{StructuredToolBase: meta.StructuredToolBase, FilePath: filePath, Search: search, Replace: replace, CreateBackup: createBackup}
		te.commandText = fmt.Sprintf("replace file content in %s", filePath)
	case "restart_service":
		service, err := stringArg(tc.Arguments, "service")
		if err != nil {
			return nil, err
		}
		te.request = RestartServiceRequest{StructuredToolBase: meta.StructuredToolBase, Service: service}
		te.commandText = fmt.Sprintf("restart service: %s", service)
	case "execute_ssh":
		command, err := stringArg(tc.Arguments, "command")
		if err != nil {
			return nil, err
		}
		cwd := optionalStringArg(tc.Arguments, "cwd")
		te.request = ToolCallRequest{Host: meta.Host, Port: meta.Port, User: meta.User, Password: meta.Password, Command: command, Cwd: cwd, Timeout: meta.Timeout}
		te.commandText = command
	default:
		return nil, fmt.Errorf("unknown tool: %s", tc.Name)
	}
	return te, nil
}

type toolMeta struct {
	StructuredToolBase
	Description string
	Thought     string
	RiskLevel   string
	RiskReason  string
}

func extractToolMeta(args map[string]interface{}, defaultHost string, creds repair.AgentSSHCreds) (toolMeta, error) {
	host := optionalStringArg(args, "host")
	if host == "" {
		host = defaultHost
	}
	user := optionalStringArg(args, "user")
	if user == "" {
		user = creds.User
	}
	password := optionalStringArg(args, "password")
	if password == "" {
		password = creds.Password
	}
	if host == "" || user == "" || password == "" {
		return toolMeta{}, fmt.Errorf("host, user and password are required")
	}
	return toolMeta{
		StructuredToolBase: StructuredToolBase{Host: host, Port: intArg(args, "port", creds.Port), User: user, Password: password, Timeout: intArg(args, "timeout", 30)},
		Description:        optionalStringArg(args, "description"),
		Thought:            optionalStringArg(args, "thought"),
		RiskLevel:          strings.ToLower(optionalStringArg(args, "risk_level")),
		RiskReason:         optionalStringArg(args, "risk_reason"),
	}, nil
}

func executeStructuredTool(client *MCPServerClient, exec *toolExecution) (*ToolCallResult, error) {
	switch req := exec.request.(type) {
	case ServiceResourceRequest:
		return client.ReadServiceResource(req)
	case KBRequest:
		return client.ReadKnowledgeBase(req)
	case KBWriteRequest:
		return client.WriteKnowledgeBase(req)
	case ServiceStatusRequest:
		return client.InspectServiceStatus(req)
	case ServiceLogsRequest:
		return client.InspectServiceLogs(req)
	case FileSnippetRequest:
		return client.InspectFileSnippet(req)
	case NginxConfigValidateRequest:
		return client.ValidateNginxConfig(req)
	case FileReplaceRequest:
		return client.ReplaceFileContent(req)
	case RestartServiceRequest:
		return client.RestartService(req)
	case ToolCallRequest:
		return client.ExecuteSSH(req)
	default:
		return nil, fmt.Errorf("unsupported tool request type")
	}
}

func publishSessionProgress(sessionMapper *mapper.RepairSessionMapper, hub *repair.StreamHub, sessionID, status string, total, completed int) {
	_ = sessionMapper.UpdateFields(sessionID, map[string]interface{}{"status": status, "total_actions": total, "completed_actions": completed})
	hub.Publish(repair.SSEEvent{Type: repair.EventSessionUpdate, SessionID: sessionID, Payload: repair.SessionUpdatePayload{Status: status, CompletedActions: completed, TotalActions: total}})
}

func countCompletedActions(actionMapper *mapper.RepairActionMapper, sessionID string) int {
	actions, err := actionMapper.GetBySessionID(sessionID)
	if err != nil {
		return 0
	}
	completed := 0
	for _, action := range actions {
		if action.Status == "success" {
			completed++
		}
	}
	return completed
}

func extractTimeout(req interface{}) int {
	switch v := req.(type) {
	case ServiceStatusRequest:
		return v.Timeout
	case ServiceLogsRequest:
		return v.Timeout
	case FileSnippetRequest:
		return v.Timeout
	case NginxConfigValidateRequest:
		return v.Timeout
	case FileReplaceRequest:
		return v.Timeout
	case RestartServiceRequest:
		return v.Timeout
	case ToolCallRequest:
		return v.Timeout
	default:
		return 30
	}
}

func buildAnalysisSummary(report reportConclusion) string {
	var parts []string
	if report.Summary != "" {
		parts = append(parts, "Summary: "+report.Summary)
	}
	if len(report.Symptoms) > 0 {
		parts = append(parts, "Symptoms: "+strings.Join(report.Symptoms, "; "))
	}
	if len(report.Evidence) > 0 {
		parts = append(parts, "Evidence: "+strings.Join(report.Evidence, "; "))
	}
	if report.FixSummary != "" {
		parts = append(parts, "Fix: "+report.FixSummary)
	}
	if report.Recommendation != "" {
		parts = append(parts, "Recommendation: "+report.Recommendation)
	}
	if len(report.NextSteps) > 0 {
		parts = append(parts, "Next steps: "+strings.Join(report.NextSteps, "; "))
	}
	return strings.Join(parts, "\n")
}

func parseReportConclusion(data []byte) reportConclusion {
	var c reportConclusion
	if err := json.Unmarshal(data, &c); err == nil {
		return c
	}
	return reportConclusion{Summary: string(data), RootCause: "see report", Severity: "medium"}
}

func stringArg(args map[string]interface{}, key string) (string, error) {
	value := optionalStringArg(args, key)
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func optionalStringArg(args map[string]interface{}, key string) string {
	value, _ := args[key].(string)
	return strings.TrimSpace(value)
}

func intArg(args map[string]interface{}, key string, fallback int) int {
	if raw, ok := args[key]; ok {
		switch value := raw.(type) {
		case float64:
			return int(value)
		case int:
			return value
		case string:
			if parsed, err := strconv.Atoi(value); err == nil {
				return parsed
			}
		}
	}
	if fallback > 0 {
		return fallback
	}
	return 22
}

func boolArg(args map[string]interface{}, key string, fallback bool) bool {
	if raw, ok := args[key]; ok {
		if value, ok := raw.(bool); ok {
			return value
		}
	}
	return fallback
}

func sameFile(left, right string) bool {
	l := strings.TrimSpace(strings.ReplaceAll(left, "\\", "/"))
	r := strings.TrimSpace(strings.ReplaceAll(right, "\\", "/"))
	return strings.EqualFold(l, r)
}

func parseValidationBlocker(output string) validationBlocker {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "nginx: [emerg]") {
			blocker := validationBlocker{rawOutput: output, summary: line}
			if idx := strings.LastIndex(line, ":"); idx != -1 && idx < len(line)-1 {
				lineNo := strings.TrimSpace(line[idx+1:])
				if _, err := strconv.Atoi(lineNo); err == nil {
					blocker.line = lineNo
				}
			}
			return blocker
		}
	}
	return validationBlocker{rawOutput: output, summary: strings.TrimSpace(output)}
}
