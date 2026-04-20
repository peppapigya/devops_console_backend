package alerting

import (
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/services/repair"
	"devops-console-backend/internal/types"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/feishu"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Orchestrator struct {
	cfg            configs.AlertingConfig
	feishuCfg      configs.FeiShuConfig
	dispatchMapper *mapper.AlertDispatchMapper
	sessionMapper  *mapper.RepairSessionMapper
	actionMapper   *mapper.RepairActionMapper
	assetMapper    *mapper.AssetHostMapper
	sessionSvc     *repair.SessionService
	agentCfg       repair.AgentLoopCfg
	notifier       feishu.Notifier
	runFn          repair.AgentRunFunc
}

func NewOrchestratorFromDB(db *gorm.DB, runFn repair.AgentRunFunc) *Orchestrator {
	alertingCfg := configs.GetAlertingConfig()
	mcpCfg := configs.GetAiConfig().MCPConfig
	mcpURL := mcpCfg.Url
	if mcpURL == "" {
		mcpURL = "http://127.0.0.1:8080"
	}

	sessionMapper := mapper.NewRepairSessionMapper(db)
	msgMapper := mapper.NewSessionMessageMapper(db)
	actionMapper := mapper.NewRepairActionMapper(db)
	eventMapper := mapper.NewRepairSessionEventMapper(db)

	return &Orchestrator{
		cfg:            alertingCfg,
		feishuCfg:      configs.GetConfig().FeiShuConfig,
		dispatchMapper: mapper.NewAlertDispatchMapper(db),
		sessionMapper:  sessionMapper,
		actionMapper:   actionMapper,
		assetMapper:    mapper.NewAssetHostMapper(db),
		sessionSvc:     repair.NewSessionService(sessionMapper, msgMapper, actionMapper, eventMapper, mcpURL),
		agentCfg: repair.AgentLoopCfg{
			LLMAPIKey:    mcpCfg.LLMApiKey,
			LLMBaseURL:   mcpCfg.LLMBaseUrl,
			LLMModel:     mcpCfg.LLMModel,
			MCPServerURL: mcpURL,
			MCPToken:     mcpCfg.Token,
			MaxRounds:    mcpCfg.MaxRounds,
		},
		notifier: feishu.NewClient(
			configs.GetConfig().FeiShuConfig.WebHookUrl,
			configs.GetConfig().FeiShuConfig.Secret,
			configs.GetConfig().FeiShuConfig.AppId,
			configs.GetConfig().FeiShuConfig.AppSecret,
			configs.GetConfig().FeiShuConfig.Enabled,
		),
		runFn: runFn,
	}
}

func (o *Orchestrator) IsEnabled() bool {
	return o.cfg.Alertmanager.Enabled
}

func (o *Orchestrator) AuthHeader() string {
	if strings.TrimSpace(o.cfg.Alertmanager.AuthHeader) != "" {
		return o.cfg.Alertmanager.AuthHeader
	}
	return "X-Alertmanager-Token"
}

func (o *Orchestrator) WebhookSecret() string {
	return o.cfg.Alertmanager.WebhookSecret
}

func (o *Orchestrator) HandleWebhook(payload types.AlertmanagerWebhookPayload) ([]types.AlertDispatchResult, error) {
	results := make([]types.AlertDispatchResult, 0, len(payload.Alerts))
	for _, alert := range payload.Alerts {
		normalized := normalizeAlert(payload, alert)
		result, err := o.handleAlert(normalized, alert)
		if err != nil {
			results = append(results, types.AlertDispatchResult{
				Fingerprint: normalized.Fingerprint,
				AlertStatus: normalized.AlertStatus,
				Dispatch:    "failed",
				Message:     err.Error(),
			})
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

func (o *Orchestrator) handleAlert(alert types.NormalizedAlert, rawAlert types.AlertmanagerAlert) (types.AlertDispatchResult, error) {
	dispatch, err := o.dispatchMapper.GetByFingerprint(alert.Fingerprint)
	if err != nil && err != gorm.ErrRecordNotFound {
		return types.AlertDispatchResult{}, err
	}

	rawPayloadBytes, _ := json.Marshal(rawAlert)
	rawPayload := string(rawPayloadBytes)

	if alert.AlertStatus == "resolved" {
		if err == gorm.ErrRecordNotFound {
			now := time.Now()
			dispatch = &model.AlertDispatch{
				Fingerprint: alert.Fingerprint,
				Status:      "resolved",
				AlertStatus: "resolved",
				AlertName:   alert.AlertName,
				AlertKey:    alert.AlertKey,
				LogEventID:  alert.LogEvent.ID,
				Host:        alert.Host,
				Service:     alert.Service,
				Level:       alert.Level,
				Message:     alert.Message,
				RawPayload:  rawPayload,
				CreatedAt:   &now,
				UpdatedAt:   &now,
			}
			if createErr := o.dispatchMapper.Create(dispatch); createErr != nil {
				return types.AlertDispatchResult{}, createErr
			}
		} else {
			now := time.Now()
			fields := map[string]interface{}{
				"status":       "resolved",
				"alert_status": "resolved",
				"message":      alert.Message,
				"raw_payload":  rawPayload,
				"updated_at":   &now,
			}
			if o.cfg.FeiShu.NotifyOnResolved && dispatch.ResolvedNotifiedAt == nil {
				_ = o.notifyResolved(dispatch, alert)
				fields["resolved_notified_at"] = &now
			}
			if updateErr := o.dispatchMapper.UpdateFields(dispatch.ID, fields); updateErr != nil {
				return types.AlertDispatchResult{}, updateErr
			}
		}
		return types.AlertDispatchResult{
			Fingerprint: alert.Fingerprint,
			AlertStatus: alert.AlertStatus,
			Dispatch:    "resolved",
			SessionID:   dispatch.SessionID,
			Message:     "resolved alert recorded",
		}, nil
	}

	if err == nil && dispatch.AlertStatus == "firing" && dispatch.SessionID != "" &&
		dispatch.Status != "resolved" && dispatch.Status != "completed" && dispatch.Status != "failed" {
		return types.AlertDispatchResult{
			Fingerprint: alert.Fingerprint,
			AlertStatus: alert.AlertStatus,
			Dispatch:    "duplicate",
			SessionID:   dispatch.SessionID,
			Message:     "alert already dispatched",
		}, nil
	}

	creds, mappingNote := o.resolveCredentials(alert)
	createReq := repair.CreateRequest{
		LogEvent: alert.LogEvent,
		SSHUser:  creds.User,
		SSHPass:  creds.Password,
		SSHPort:  creds.Port,
		Operator: "alertmanager",
	}

	sessionID := ""
	if o.cfg.Analysis.AutoCreateSession {
		sessionID, err = o.sessionSvc.CreateSession(createReq)
		if err != nil {
			return types.AlertDispatchResult{}, err
		}
		o.sessionSvc.StartAsync(sessionID, createReq, o.agentCfg, o.runFn)
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":       "analyzing",
		"alert_status": "firing",
		"alert_name":   alert.AlertName,
		"alert_key":    alert.AlertKey,
		"session_id":   sessionID,
		"log_event_id": alert.LogEvent.ID,
		"host":         alert.Host,
		"service":      alert.Service,
		"level":        alert.Level,
		"message":      alert.Message,
		"raw_payload":  rawPayload,
		"updated_at":   &now,
	}
	if err == gorm.ErrRecordNotFound {
		dispatch = &model.AlertDispatch{
			Fingerprint: alert.Fingerprint,
			Status:      "analyzing",
			AlertStatus: "firing",
			AlertName:   alert.AlertName,
			AlertKey:    alert.AlertKey,
			SessionID:   sessionID,
			LogEventID:  alert.LogEvent.ID,
			Host:        alert.Host,
			Service:     alert.Service,
			Level:       alert.Level,
			Message:     alert.Message,
			RawPayload:  rawPayload,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		}
		if createErr := o.dispatchMapper.Create(dispatch); createErr != nil {
			return types.AlertDispatchResult{}, createErr
		}
	} else {
		if updateErr := o.dispatchMapper.UpdateFields(dispatch.ID, fields); updateErr != nil {
			return types.AlertDispatchResult{}, updateErr
		}
	}

	if o.cfg.FeiShu.NotifyOnFiring {
		_ = o.notifyFiring(dispatch.ID, sessionID, alert, mappingNote)
	}
	if sessionID != "" {
		go o.watchSession(dispatch.ID, sessionID, alert)
	}

	return types.AlertDispatchResult{
		Fingerprint: alert.Fingerprint,
		AlertStatus: alert.AlertStatus,
		Dispatch:    "created",
		SessionID:   sessionID,
		Message:     mappingNote,
	}, nil
}

func (o *Orchestrator) resolveCredentials(alert types.NormalizedAlert) (repair.AgentSSHCreds, string) {
	if !o.cfg.AssetMapping.Enabled {
		return repair.AgentSSHCreds{}, "asset mapping disabled"
	}

	var host model.AssetHost
	query := o.assetMapper.DB.Model(&model.AssetHost{}).Where("deleted_at IS NULL")
	hostValue := strings.TrimSpace(alert.Host)
	if hostValue != "" {
		err := query.Where("ip = ? OR name = ?", hostValue, hostValue).Order("id DESC").First(&host).Error
		if err == nil {
			password, _ := o.assetMapper.GetDecryptedPassword(host.ID)
			return repair.AgentSSHCreds{
				User:     host.Username,
				Password: password,
				Port:     int(host.Port),
			}, fmt.Sprintf("mapped from asset host %s", host.Name)
		}
	}

	if o.cfg.AssetMapping.AllowNoCredential {
		return repair.AgentSSHCreds{}, "no asset credential matched; fallback to no-credential analysis"
	}
	return repair.AgentSSHCreds{}, "no asset credential matched"
}

func (o *Orchestrator) watchSession(dispatchID uint64, sessionID string, alert types.NormalizedAlert) {
	interval := time.Duration(o.cfg.Analysis.PollIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(o.cfg.Analysis.PollTimeoutSeconds) * time.Second)
	if o.cfg.Analysis.PollTimeoutSeconds <= 0 {
		deadline = time.Now().Add(15 * time.Minute)
	}

	for time.Now().Before(deadline) {
		session, err := o.sessionMapper.GetByID(sessionID)
		if err == nil {
			notifyStatus := session.Status == "waiting_approval" || session.Status == "completed" || session.Status == "failed" || session.Status == "paused"
			dispatch, dispatchErr := o.dispatchMapper.GetByFingerprint(alert.Fingerprint)
			if dispatchErr == nil {
				if notifyStatus && dispatch.LastSessionStatus != session.Status && o.cfg.FeiShu.NotifyOnResult {
					_ = o.notifyResult(session, dispatch, alert)
					now := time.Now()
					_ = o.dispatchMapper.UpdateFields(dispatchID, map[string]interface{}{
						"last_session_status": session.Status,
						"result_notified_at":  &now,
						"status":              session.Status,
						"updated_at":          &now,
					})
				} else {
					now := time.Now()
					_ = o.dispatchMapper.UpdateFields(dispatchID, map[string]interface{}{
						"status":              session.Status,
						"last_session_status": session.Status,
						"updated_at":          &now,
					})
				}
			}
			if session.Status == "completed" || session.Status == "failed" || session.Status == "paused" {
				return
			}
		}
		time.Sleep(interval)
	}
}

func normalizeAlert(payload types.AlertmanagerWebhookPayload, alert types.AlertmanagerAlert) types.NormalizedAlert {
	labels := cloneStringMap(payload.CommonLabels)
	for key, value := range alert.Labels {
		labels[key] = value
	}
	annotations := cloneStringMap(payload.CommonAnnotations)
	for key, value := range alert.Annotations {
		annotations[key] = value
	}

	host := firstNonEmpty(
		labels["host"],
		labels["instance"],
		labels["node"],
		labels["pod_ip"],
	)
	host = strings.TrimSpace(strings.Split(host, ":")[0])
	service := firstNonEmpty(labels["service"], labels["job"], labels["pod"], labels["container"], labels["namespace"])
	level := normalizeLevel(firstNonEmpty(labels["severity"], labels["level"], payload.Status, alert.Status))
	alertName := firstNonEmpty(labels["alertname"], "PrometheusAlert")
	message := firstNonEmpty(
		annotations["summary"],
		annotations["description"],
		annotations["message"],
		alertName,
	)
	source := firstNonEmpty(service, alertName, payload.Receiver, "alertmanager")
	fingerprint := strings.TrimSpace(alert.Fingerprint)
	if fingerprint == "" {
		fingerprint = fmt.Sprintf("%s:%s:%s:%s", alertName, host, service, message)
	}
	alertKey := fmt.Sprintf("%s:%s:%s", alertName, host, service)
	metadata := map[string]interface{}{
		"group_key":     payload.GroupKey,
		"receiver":      payload.Receiver,
		"external_url":  payload.ExternalURL,
		"generator_url": alert.GeneratorURL,
		"labels":        labels,
		"annotations":   annotations,
	}

	return types.NormalizedAlert{
		Fingerprint: fingerprint,
		AlertName:   alertName,
		AlertKey:    alertKey,
		AlertStatus: strings.ToLower(firstNonEmpty(alert.Status, payload.Status, "firing")),
		Source:      source,
		Host:        host,
		Service:     service,
		Level:       level,
		Message:     message,
		Labels:      labels,
		Annotations: annotations,
		Metadata:    metadata,
		LogEvent: types.LogEvent{
			ID:        fmt.Sprintf("alert-%s", fingerprint),
			Timestamp: chooseAlertTime(alert.StartsAt),
			Source:    source,
			Level:     level,
			Message:   message,
			Host:      host,
			Service:   service,
			Metadata:  metadata,
		},
	}
}

func (o *Orchestrator) notifyFiring(dispatchID uint64, sessionID string, alert types.NormalizedAlert, mappingNote string) error {
	title := fmt.Sprintf("Prometheus 告警已接收: %s", alert.AlertName)
	content := strings.Join([]string{
		fmt.Sprintf("**状态**: `%s`", alert.AlertStatus),
		fmt.Sprintf("**级别**: `%s`", alert.Level),
		fmt.Sprintf("**主机/服务**: `%s` / `%s`", emptyAsDash(alert.Host), emptyAsDash(alert.Service)),
		fmt.Sprintf("**摘要**: %s", alert.Message),
		fmt.Sprintf("**分析会话**: `%s`", emptyAsDash(sessionID)),
		fmt.Sprintf("**资产映射**: %s", mappingNote),
		buildRcaLinkMarkdown(o.cfg.ConsoleBaseURL, sessionID),
	}, "\n")
	if err := o.notifier.SendCard(title, content); err != nil {
		return err
	}
	now := time.Now()
	return o.dispatchMapper.UpdateFields(dispatchID, map[string]interface{}{
		"firing_notified_at": &now,
		"updated_at":         &now,
	})
}

func (o *Orchestrator) notifyResolved(dispatch *model.AlertDispatch, alert types.NormalizedAlert) error {
	title := fmt.Sprintf("Prometheus 告警已恢复: %s", alert.AlertName)
	content := strings.Join([]string{
		fmt.Sprintf("**状态**: `%s`", alert.AlertStatus),
		fmt.Sprintf("**主机/服务**: `%s` / `%s`", emptyAsDash(alert.Host), emptyAsDash(alert.Service)),
		fmt.Sprintf("**摘要**: %s", alert.Message),
		fmt.Sprintf("**分析会话**: `%s`", emptyAsDash(dispatch.SessionID)),
		buildRcaLinkMarkdown(o.cfg.ConsoleBaseURL, dispatch.SessionID),
	}, "\n")
	return o.notifier.SendCard(title, content)
}

func (o *Orchestrator) notifyResult(session *model.RepairSession, dispatch *model.AlertDispatch, alert types.NormalizedAlert) error {
	title := fmt.Sprintf("MCP 分析结果: %s", alert.AlertName)
	actions, _ := o.actionMapper.GetBySessionID(session.ID)
	waitingAction := ""
	for _, action := range actions {
		if action.Status == "waiting_confirm" {
			waitingAction = action.Description
			if action.RiskReason != "" {
				waitingAction += "；风险说明: " + action.RiskReason
			}
			break
		}
	}

	analysis := session.Analysis
	if strings.TrimSpace(analysis) == "" && waitingAction != "" {
		analysis = "分析流程已推进到人工审批阶段。"
	}
	if strings.TrimSpace(analysis) == "" {
		analysis = "分析仍在进行中，请在系统内查看实时进度。"
	}

	lines := []string{
		fmt.Sprintf("**会话状态**: `%s`", session.Status),
		fmt.Sprintf("**严重度**: `%s`", emptyAsDash(session.Severity)),
		fmt.Sprintf("**置信度**: `%d%%`", int(session.Confidence*100)),
		fmt.Sprintf("**根因**: %s", emptyAsDash(session.RootCause)),
		fmt.Sprintf("**分析摘要**: %s", analysis),
	}
	if waitingAction != "" {
		lines = append(lines, fmt.Sprintf("**待审批动作**: %s", waitingAction))
		lines = append(lines, "**处理建议**: 需要在系统 RCA 页面完成人工审批后再继续执行。")
	}
	lines = append(lines, buildRcaLinkMarkdown(o.cfg.ConsoleBaseURL, session.ID))
	return o.notifier.SendCard(title, strings.Join(lines, "\n"))
}

func buildRcaLinkMarkdown(baseURL, sessionID string) string {
	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(sessionID) == "" {
		return ""
	}
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return fmt.Sprintf("**查看详情**: [打开 RCA 页面](%s/monitor/rca?session_id=%s)", base, sessionID)
}

func chooseAlertTime(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now()
	}
	return ts
}

func normalizeLevel(raw string) string {
	value := strings.ToUpper(strings.TrimSpace(raw))
	switch value {
	case "CRITICAL", "P1":
		return "CRITICAL"
	case "HIGH", "WARNING", "WARN", "P2":
		return "WARN"
	case "INFO", "LOW", "P3", "P4":
		return "INFO"
	case "RESOLVED":
		return "INFO"
	default:
		if value == "" {
			return "ERROR"
		}
		return value
	}
}

func cloneStringMap(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func emptyAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
