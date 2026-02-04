package node

import (
	"devops-console-backend/internal/controllers/k8s/node"

	"github.com/gin-gonic/gin"
)

// NodeRoute Node路由
type NodeRoute struct {
	controller *node.NodeController
}

// NewNodeRoute 创建Node路由实例
func NewNodeRoute() *NodeRoute {
	return &NodeRoute{
		controller: node.NewNodeController(),
	}
}

// RegisterSubRouter 注册子路由
func (r *NodeRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	nodeGroup := apiGroup.Group("/k8s/node")
	{
		nodeGroup.GET("/list", r.controller.GetNodeList)
		nodeGroup.GET("/detail/:nodeName", r.controller.GetNodeDetail)
		nodeGroup.POST("/:nodeName/cordon", r.controller.CordonNode)
		nodeGroup.POST("/:nodeName/uncordon", r.controller.UncordonNode)
		nodeGroup.POST("/:nodeName/drain", r.controller.DrainNode)
		nodeGroup.POST("/:nodeName/labels", r.controller.AddNodeLabel)
		nodeGroup.DELETE("/:nodeName/labels/:labelKey", r.controller.RemoveNodeLabel)
	}
}
