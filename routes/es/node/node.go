package node

import (
	node2 "devops-console-backend/controllers/es/node"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册集群相关路由
func RegisterRoutes(nodeGroup *gin.RouterGroup) {
	nodeGroup.GET("/ClusterNodeInfo", node2.ClusterNodeInfo)
	nodeGroup.GET("/ClusterNodeStats", node2.ClusterNodeStats)
}

func RegisterSubRouter(g *gin.RouterGroup) {
	nodeGroup := g.Group("/node")
	RegisterRoutes(nodeGroup)
}
