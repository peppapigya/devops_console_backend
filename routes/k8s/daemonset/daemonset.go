package daemonset

import (
	"devops-console-backend/controllers/k8s/daemonset"
	"github.com/gin-gonic/gin"
)

// DaemonSetRoute DaemonSet路由
type DaemonSetRoute struct {
	controller *daemonset.DaemonSetController
}

// NewDaemonSetRoute 创建DaemonSet路由实例
func NewDaemonSetRoute() *DaemonSetRoute {
	return &DaemonSetRoute{
		controller: daemonset.NewDaemonSetController(),
	}
}

// RegisterSubRouter 注册子路由
func (r *DaemonSetRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	daemonSetGroup := apiGroup.Group("/k8s/daemonset")
	{
		daemonSetGroup.GET("/detail/:namespace/:daemonSetName", r.controller.GetDaemonSetDetail)
		daemonSetGroup.GET("/list/:namespace", r.controller.GetDaemonSetList)
		daemonSetGroup.POST("/create/:namespace", r.controller.CreateDaemonSet)
		daemonSetGroup.PUT("/update/:namespace/:daemonSetName", r.controller.UpdateDaemonSet)
		daemonSetGroup.DELETE("/delete/:namespace/:daemonSetName", r.controller.DeleteDaemonSet)
	}
}