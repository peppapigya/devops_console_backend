package mcp

import (
	"devops-console-backend/internal/agent"
	monitorCtrl "devops-console-backend/internal/controllers/monitor"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/services/repair"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type MCPAgentController struct {
	sessionSvc *repair.SessionService
	agentCfg   repair.AgentLoopCfg
}

func NewMCPAgentController(sessionSvc *repair.SessionService, agentCfg repair.AgentLoopCfg) *MCPAgentController {
	return &MCPAgentController{sessionSvc: sessionSvc, agentCfg: agentCfg}
}

func NewMCPAgentControllerFromDB() *MCPAgentController {
	gdb := configs.GORMDB
	sessionMapper := mapper.NewRepairSessionMapper(gdb)
	msgMapper := mapper.NewSessionMessageMapper(gdb)
	actionMapper := mapper.NewRepairActionMapper(gdb)
	eventMapper := mapper.NewRepairSessionEventMapper(gdb)

	mcpCfg := configs.GetAiConfig().MCPConfig
	mcpURL := mcpCfg.Url
	if mcpURL == "" {
		mcpURL = "http://127.0.0.1:8080"
	}

	svc := repair.NewSessionService(sessionMapper, msgMapper, actionMapper, eventMapper, mcpURL)
	agentCfg := repair.AgentLoopCfg{
		LLMAPIKey:    mcpCfg.LLMApiKey,
		LLMBaseURL:   mcpCfg.LLMBaseUrl,
		LLMModel:     mcpCfg.LLMModel,
		MCPServerURL: mcpURL,
		MCPToken:     mcpCfg.Token,
		MaxRounds:    mcpCfg.MaxRounds,
	}
	return &MCPAgentController{sessionSvc: svc, agentCfg: agentCfg}
}

func (ctrl *MCPAgentController) CreateSession(c *gin.Context) {
	helper := utils.NewResponseHelper(c)
	aiCfg := configs.GetAiConfig().MCPConfig
	if !aiCfg.Enabled {
		helper.BadRequest("AI feature is disabled")
		return
	}

	var req repair.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("invalid request: " + err.Error())
		return
	}
	if req.Operator == "" {
		req.Operator = c.GetString("username")
	}

	sessionID, err := ctrl.sessionSvc.CreateSession(req)
	if err != nil {
		helper.InternalError("create session failed: " + err.Error())
		return
	}

	ctrl.sessionSvc.StartAsync(sessionID, req, ctrl.agentCfg, agent.RunAgentLoop)
	monitorCtrl.RepairSessionsTotal.WithLabelValues("created").Inc()
	helper.Success("success", map[string]interface{}{"session_id": sessionID})
}

func (ctrl *MCPAgentController) StreamSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "session_id is required"})
		return
	}
	sinceID, _ := strconv.ParseUint(c.DefaultQuery("since_id", "0"), 10, 64)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	hub := repair.GetHub()
	flusher, canFlush := c.Writer.(http.Flusher)
	if replayLines, _, err := hub.Replay(sessionID, sinceID); err == nil {
		for _, line := range replayLines {
			_, _ = io.WriteString(c.Writer, line)
		}
		if canFlush {
			flusher.Flush()
		}
	}

	ch, cancel := hub.Subscribe(sessionID)
	defer cancel()
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			_, _ = io.WriteString(c.Writer, msg)
			if canFlush {
				flusher.Flush()
			}
			if msg == "event: close\ndata: {}\n\n" {
				return
			}
		case <-ticker.C:
			_, _ = io.WriteString(c.Writer, ": heartbeat\n\n")
			if canFlush {
				flusher.Flush()
			}
		case <-c.Request.Context().Done():
			return
		}
	}
}

func (ctrl *MCPAgentController) ApproveAction(c *gin.Context) {
	helper := utils.NewResponseHelper(c)
	sessionID := c.Param("id")
	actionID, err := strconv.ParseUint(c.Param("actionId"), 10, 64)
	if err != nil {
		helper.BadRequest("invalid actionId")
		return
	}

	var body struct {
		Approved bool `json:"approved"`
	}
	_ = c.ShouldBindJSON(&body)
	gdb := configs.GORMDB
	actionMapper := mapper.NewRepairActionMapper(gdb)
	now := time.Now()
	approver := c.GetString("username")
	updateFields := map[string]interface{}{
		"approved_by": approver,
		"approved_at": &now,
	}
	if body.Approved {
		updateFields["status"] = "approved"
	} else {
		updateFields["status"] = "skipped"
	}
	_ = actionMapper.UpdateFields(actionID, updateFields)
	if ok := repair.ConfirmAction(sessionID, actionID, body.Approved); !ok {
		helper.NotFound("waiting action not found or already expired")
		return
	}
	decision := "rejected"
	if body.Approved {
		decision = "approved"
	}
	monitorCtrl.RepairApprovalsTotal.WithLabelValues(decision).Inc()
	helper.Success("success")
}

func (ctrl *MCPAgentController) PauseSession(c *gin.Context) {
	helper := utils.NewResponseHelper(c)
	sessionID := c.Param("id")
	gdb := configs.GORMDB
	actionMapper := mapper.NewRepairActionMapper(gdb)
	actions, _ := actionMapper.GetBySessionID(sessionID)
	for _, a := range actions {
		if a.Status == "waiting_confirm" {
			repair.ConfirmAction(sessionID, a.ID, false)
		}
	}
	sessionMapper := mapper.NewRepairSessionMapper(gdb)
	_ = sessionMapper.UpdateStatus(sessionID, "paused")
	helper.Success("pause requested")
}

func (ctrl *MCPAgentController) ListSessions(c *gin.Context) {
	helper := utils.NewResponseHelper(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	gdb := configs.GORMDB
	sessionMapper := mapper.NewRepairSessionMapper(gdb)
	total, list, err := sessionMapper.ListPage(page, pageSize)
	if err != nil {
		helper.InternalError("query failed: " + err.Error())
		return
	}
	normalizeSessionListForResponse(list)
	helper.Success("success", map[string]interface{}{"list": list, "total": total})
}

func (ctrl *MCPAgentController) GetSession(c *gin.Context) {
	helper := utils.NewResponseHelper(c)
	sessionID := c.Param("id")
	gdb := configs.GORMDB
	sessionMapper := mapper.NewRepairSessionMapper(gdb)
	actionMapper := mapper.NewRepairActionMapper(gdb)
	session, err := sessionMapper.GetByID(sessionID)
	if err != nil {
		helper.NotFound("session not found")
		return
	}
	normalizeSessionForResponse(session)
	actions, _ := actionMapper.GetBySessionID(sessionID)
	helper.Success("success", map[string]interface{}{"session": session, "actions": actions})
}
