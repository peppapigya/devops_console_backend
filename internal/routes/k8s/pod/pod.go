package pod

import (
	"devops-console-backend/internal/controllers/k8s/pod"

	"github.com/gin-gonic/gin"
)

// PodRoute Pod路由
type PodRoute struct {
	controller *pod.PodController
}

// NewPodRoute 创建Pod路由实例
func NewPodRoute() *PodRoute {
	return &PodRoute{
		controller: pod.NewPodController(),
	}
}

// RegisterSubRouter 注册子路由
func (r *PodRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	podGroup := apiGroup.Group("/k8s/pod")
	{
		podGroup.GET("/detail/:namespace/:podname", r.controller.GetPodDetail)
		podGroup.GET("/list/:namespace", r.controller.GetPodList)
		podGroup.GET("/list/all", r.controller.GetPodList)
		podGroup.GET("/events/:namespace/:podname", r.controller.GetPodEvents)
		podGroup.GET("/logs/:namespace/:podname", r.controller.GetPodLogs)
		podGroup.POST("/create", r.controller.CreatePod)
		podGroup.PUT("/update", r.controller.UpdatePod)
		podGroup.DELETE("/delete/:namespace/:podname", r.controller.DeletePod)
	}
}
