package mcp

import (
	"devops-console-backend/internal/mcp"
	"devops-console-backend/internal/types"
	"devops-console-backend/pkg/configs"
	"time"

	"github.com/gin-gonic/gin"
)

type MCPAgentController struct {
}

func NewMCPAgentController() *MCPAgentController {
	return &MCPAgentController{}
}

func (agent *MCPAgentController) Repair(c *gin.Context) {
	mcpConfig := configs.GetAiConfig().MCPConfig
	if !mcpConfig.Enabled {
		c.JSON(400, gin.H{"error": "ai功能未启用"})
		return
	}
	var logEvent *types.LogEvent

	if err := c.ShouldBindBodyWithJSON(&logEvent); err != nil {
		c.JSON(400, gin.H{"error": "参数解析失败"})
		return
	}

	client := mcp.NewMCPClient("http://127.0.0.1:8080", "")

	agentContext := &types.AgentContext{
		LogEvent:      *logEvent,
		Timestamp:     time.Now(),
		AttemptCount:  1,
		PreviousFixes: make([]types.FixResult, 0),
	}
	for i := 0; i < mcpConfig.MaxRetries; i++ {
		suggestion, err := client.GetSuggestion(agentContext)
		if err != nil {
			c.JSON(400, gin.H{"error": "获取ai建议失败"})
			break
		}
		// 去执行对应的修复逻辑
		work, err := agent.executeWork(suggestion)
		if err != nil {
			c.JSON(400, gin.H{"error": "执行修复逻辑失败"})
			break
		}
		agentContext.PreviousFixes = append(agentContext.PreviousFixes, work)
		if work.Success {
			break
		}
	}
	c.JSON(200, gin.H{"data": agentContext.PreviousFixes})
}

func (agent *MCPAgentController) executeWork(suggestion *types.FixSuggestion) (types.FixResult, error) {
	return types.FixResult{}, nil
}
