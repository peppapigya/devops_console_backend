package mcp

import (
	"devops-console-backend/internal/agent"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/services/repair"
	"devops-console-backend/pkg/configs"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MCPAgentController 管理根因分析与修复 Session 的 HTTP 接口
type MCPAgentController struct {
	sessionSvc *repair.SessionService
	agentCfg   repair.AgentLoopCfg
}

func NewMCPAgentController(sessionSvc *repair.SessionService, agentCfg repair.AgentLoopCfg) *MCPAgentController {
	return &MCPAgentController{sessionSvc: sessionSvc, agentCfg: agentCfg}
}

// NewMCPAgentControllerFromDB 从全局 DB 连接构建 Controller（供路由层使用）
func NewMCPAgentControllerFromDB() *MCPAgentController {
	gdb := configs.GORMDB
	sessionMapper := mapper.NewRepairSessionMapper(gdb)
	msgMapper := mapper.NewSessionMessageMapper(gdb)
	actionMapper := mapper.NewRepairActionMapper(gdb)

	mcpCfg := configs.GetAiConfig().MCPConfig
	mcpURL := mcpCfg.Url
	if mcpURL == "" {
		mcpURL = "http://127.0.0.1:8080"
	}

	svc := repair.NewSessionService(sessionMapper, msgMapper, actionMapper, mcpURL)

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

// ──────────────────────────────────────────────────────────────
// POST /api/v1/mcp/session — 创建 Session，立即返回 session_id
// ──────────────────────────────────────────────────────────────
func (ctrl *MCPAgentController) CreateSession(c *gin.Context) {
	aiCfg := configs.GetAiConfig().MCPConfig
	if !aiCfg.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"message": "AI 功能未启用"})
		return
	}

	var req repair.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数解析失败: " + err.Error()})
		return
	}

	sessionID, err := ctrl.sessionSvc.CreateSession(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "创建 session 失败: " + err.Error()})
		return
	}

	// 异步启动 MCP Agent 循环（注入 agent.RunAgentLoop 函数）
	ctrl.sessionSvc.StartAsync(sessionID, req, ctrl.agentCfg, agent.RunAgentLoop)

	c.JSON(http.StatusOK, gin.H{
		"data":    gin.H{"session_id": sessionID},
		"message": "success",
	})
}

// ──────────────────────────────────────────────────────────────
// GET /api/v1/mcp/session/:id/stream — SSE 长连接实时推送
// ──────────────────────────────────────────────────────────────
func (ctrl *MCPAgentController) StreamSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "session_id 不能为空"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	hub := repair.GetHub()
	ch, cancel := hub.Subscribe(sessionID)
	defer cancel()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	flusher, canFlush := c.Writer.(http.Flusher)

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

// ──────────────────────────────────────────────────────────────
// POST /api/v1/mcp/session/:id/action/:actionId/approve
// ──────────────────────────────────────────────────────────────
func (ctrl *MCPAgentController) ApproveAction(c *gin.Context) {
	sessionID := c.Param("id")
	actionIDStr := c.Param("actionId")
	actionID, err := strconv.ParseUint(actionIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "actionId 格式错误"})
		return
	}

	var body struct {
		Approved bool `json:"approved"`
	}
	_ = c.ShouldBindJSON(&body)

	ok := repair.ConfirmAction(sessionID, actionID, body.Approved)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"message": "未找到等待确认的动作，可能已超时"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// ──────────────────────────────────────────────────────────────
// POST /api/v1/mcp/session/:id/pause
// ──────────────────────────────────────────────────────────────
func (ctrl *MCPAgentController) PauseSession(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"message": "已请求暂停"})
}

// ──────────────────────────────────────────────────────────────
// GET /api/v1/mcp/session/list — 历史 session 列表
// ──────────────────────────────────────────────────────────────
func (ctrl *MCPAgentController) ListSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	gdb := configs.GORMDB
	sessionMapper := mapper.NewRepairSessionMapper(gdb)
	total, list, err := sessionMapper.ListPage(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    gin.H{"list": list, "total": total},
		"message": "success",
	})
}

// ──────────────────────────────────────────────────────────────
// GET /api/v1/mcp/session/:id — session 详情
// ──────────────────────────────────────────────────────────────
func (ctrl *MCPAgentController) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	gdb := configs.GORMDB
	sessionMapper := mapper.NewRepairSessionMapper(gdb)
	actionMapper := mapper.NewRepairActionMapper(gdb)

	session, err := sessionMapper.GetByID(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "session 不存在"})
		return
	}
	actions, _ := actionMapper.GetBySessionID(sessionID)

	c.JSON(http.StatusOK, gin.H{
		"data":    gin.H{"session": session, "actions": actions},
		"message": "success",
	})
}
