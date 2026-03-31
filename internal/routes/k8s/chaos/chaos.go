package chaos

import (
	"devops-console-backend/internal/controllers/k8s/chaos"

	"github.com/gin-gonic/gin"
)

// ChaosRoute Chaos Mesh路由
type ChaosRoute struct {
	controller *chaos.ChaosController
}

// NewChaosRoute 创建Chaos Mesh路由实例
func NewChaosRoute() *ChaosRoute {
	return &ChaosRoute{
		controller: chaos.NewChaosController(),
	}
}

// RegisterSubRouter 注册子路由
func (r *ChaosRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	chaosGroup := apiGroup.Group("/k8s/chaos")
	{
		chaosGroup.POST("/create/:namespace", r.controller.CreateFault)
		chaosGroup.GET("/list/:namespace", r.controller.ListFaults)
		chaosGroup.GET("/get/:namespace/:name", r.controller.GetFault)
		chaosGroup.DELETE("/delete/:namespace/:name", r.controller.DeleteFault)
		chaosGroup.PUT("/pause/:namespace/:name", r.controller.PauseFault)
		chaosGroup.PUT("/resume/:namespace/:name", r.controller.ResumeFault)
		// 演练节点驱逐
		chaosGroup.GET("/nodes", r.controller.GetChaosNodes)
		chaosGroup.POST("/evict/prepare", r.controller.PrepareEviction)
		chaosGroup.POST("/evict/cleanup", r.controller.CleanupEviction)
	}
}
