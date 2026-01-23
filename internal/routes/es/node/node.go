package node

import (
	node3 "devops-console-backend/internal/controllers/es/node"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册集群相关路由
func RegisterRoutes(nodeGroup *gin.RouterGroup) {
	nodeGroup.GET("/ClusterNodeInfo", node3.ClusterNodeInfo)
	nodeGroup.GET("/ClusterNodeStats", node3.ClusterNodeStats)
}

func RegisterSubRouter(g *gin.RouterGroup) {
	nodeGroup := g.Group("/node")
	RegisterRoutes(nodeGroup)
}
