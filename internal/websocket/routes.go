package websocket

import (
	"github.com/gin-gonic/gin"
)

// RegisterWebSocketRoutes 注册WebSocket路由
func RegisterWebSocketRoutes(r *gin.Engine) {
	ws := r.Group("/ws")

	// Pod日志WebSocket
	ws.GET("/pod/:podname/logs", NewPodLogHandler().HandleWebSocket)

	// Pod终端WebSocket
	ws.GET("/pod/:podname/exec", NewPodExecHandler().HandleWebSocket)
}
