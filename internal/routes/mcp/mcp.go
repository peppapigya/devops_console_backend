package mcp

import (
	mcpCtrl "devops-console-backend/internal/controllers/mcp"
	"devops-console-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterMCPRouters 注册 MCP 根因分析路由
func RegisterMCPRouters(router *gin.RouterGroup, db *gorm.DB) {
	ctrl := mcpCtrl.NewMCPAgentControllerFromDB()

	mcpGroup := router.Group("/mcp")
	mcpGroup.Use(middlewares.Authenticate())
	{
		// 创建 session（异步触发分析）
		mcpGroup.POST("/session", ctrl.CreateSession)
		// session 列表
		mcpGroup.GET("/session/list", ctrl.ListSessions)
		// session 详情
		mcpGroup.GET("/session/:id", ctrl.GetSession)
		// SSE 流（实时推送分析过程）
		mcpGroup.GET("/session/:id/stream", ctrl.StreamSession)
		// 用户确认高风险动作
		mcpGroup.POST("/session/:id/action/:actionId/approve", ctrl.ApproveAction)
		// 暂停 session
		mcpGroup.POST("/session/:id/pause", ctrl.PauseSession)
	}
}
