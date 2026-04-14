package agent

import (
	"context"
	monitorCtrl "devops-console-backend/internal/controllers/monitor"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/services/repair"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/gopkg/util/logger"
)

const defaultMaxRounds = 10

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
	llmClient.SetSystemPrompt(systemPrompt)
	llmClient.AddUserMessage(initialUserMsg)

	now := time.Now()
	_ = msgMapper.BatchCreate([]*model.SessionMessage{
		{SessionID: sessionID, Role: "system", Content: systemPrompt, CreatedAt: &now},
		{SessionID: sessionID, Role: "user", Content: initialUserMsg, CreatedAt: &now},
	})

	actionOrder := 0
	lastActionSuccess := true
	lastActionOutput := ""
	commandAttempts := map[string]int{}

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

				if exec.riskLevel == "high" {
					_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{"status": "waiting_confirm"})
					publishSessionProgress(sessionMapper, hub, sessionID, "waiting_approval", actionOrder, actionOrder-1)
					hub.Publish(repair.SSEEvent{Type: repair.EventWaitConfirm, SessionID: sessionID, Payload: repair.WaitConfirmPayload{ActionID: action.ID, ActionOrder: actionOrder, Description: exec.description, Command: exec.commandText, RiskReason: exec.riskReason}})
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
				hub.Publish(repair.SSEEvent{Type: repair.EventActionStart, SessionID: sessionID, Payload: repair.ActionStartPayload{ActionID: action.ID, ActionOrder: actionOrder, ToolName: exec.name, Thought: exec.thought, Description: exec.description, Command: exec.commandText, Target: exec.target}})

				toolResult, toolErr := executeStructuredTool(mcpClient, exec)
				finishedAt := time.Now()
				if toolErr != nil {
					_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{"status": "failed", "error_msg": toolErr.Error(), "exit_code": -1, "duration_ms": 0, "executed_at": &finishedAt})
					monitorCtrl.RepairActionsTotal.WithLabelValues(exec.name, "failed", exec.riskLevel).Inc()
					hub.Publish(repair.SSEEvent{Type: repair.EventActionResult, SessionID: sessionID, Payload: repair.ActionResultPayload{ActionID: action.ID, Status: "failed", ErrorMsg: toolErr.Error(), ExitCode: -1}})
					llmClient.AddToolResult(tc.ID, fmt.Sprintf(`{"success":false,"error":%q,"exit_code":-1}`, toolErr.Error()))
					continue
				}

				status := "success"
				if !toolResult.Success {
					status = "failed"
				}
				lastActionSuccess = toolResult.Success
				rawActionOutput := strings.TrimSpace(toolResult.Output + "\n" + toolResult.Stderr)
				lastActionOutput = rawActionOutput
				if len(lastActionOutput) > 1200 {
					lastActionOutput = lastActionOutput[:1200] + "..."
				}
				_ = actionMapper.UpdateFields(action.ID, map[string]interface{}{"status": status, "output": rawActionOutput, "exit_code": toolResult.ExitCode, "duration_ms": int(toolResult.DurationMs), "executed_at": &finishedAt})
				monitorCtrl.RepairActionsTotal.WithLabelValues(exec.name, status, exec.riskLevel).Inc()
				monitorCtrl.RepairActionDuration.WithLabelValues(exec.name, status).Observe(float64(toolResult.DurationMs) / 1000.0)
				publishSessionProgress(sessionMapper, hub, sessionID, "executing", actionOrder, countCompletedActions(actionMapper, sessionID))
				hub.Publish(repair.SSEEvent{Type: repair.EventActionResult, SessionID: sessionID, Payload: repair.ActionResultPayload{ActionID: action.ID, Status: status, Output: toolResult.Output, ErrorMsg: toolResult.Stderr, ExitCode: toolResult.ExitCode, DurationMs: int(toolResult.DurationMs)}})

				compacted := CompactToolResult(exec.name, exec.commandText, toolResult)
				resultJSON, _ := json.Marshal(compacted)
				llmClient.AddToolResult(tc.ID, string(resultJSON))
				if exec.name == "validate_nginx_config" && toolResult.Success {
					llmClient.AddUserMessage("nginx 配置验证已成功。下一步必须调用 submit_diagnosis_report 提交最终结论，不要继续重复取证。")
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
			llmClient.AddUserMessage("不要只输出纯文本。请调用工具收集证据，或调用 submit_diagnosis_report 结束流程。并且推理描述必须中文。")
			continue
		}
	}

	return fmt.Errorf("agent reached max diagnosis rounds (%d) without completion", maxRounds)
}

func handleDiagnosisReport(tc ToolCall, sessionID string, actionMapper *mapper.RepairActionMapper, sessionMapper *mapper.RepairSessionMapper, msgMapper *mapper.SessionMessageMapper, hub *repair.StreamHub, lastActionSuccess bool, lastActionOutput string, llmClient *LLMClient) error {
	args := tc.Arguments
	fixed, _ := args["fixed"].(bool)
	if fixed && !lastActionSuccess {
		msg := fmt.Sprintf("最近一次验证/执行失败，不能提交 fixed=true。最近输出：%s", lastActionOutput)
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
