package statefulset

import (
	"devops-console-backend/internal/controllers/k8s/statefulset"

	"github.com/gin-gonic/gin"
)

// StatefulSetRoute StatefulSet路由
type StatefulSetRoute struct {
	statefulSetController *statefulset.StatefulSetController
}

// NewStatefulSetRoute 创建StatefulSet路由实例
func NewStatefulSetRoute() *StatefulSetRoute {
	return &StatefulSetRoute{
		statefulSetController: statefulset.NewStatefulSetController(),
	}
}

// RegisterSubRouter 注册子路由
func (r *StatefulSetRoute) RegisterSubRouter(rg *gin.RouterGroup) {
	// StatefulSet相关路由
	statefulSetGroup := rg.Group("/k8s/statefulset")
	{
		statefulSetGroup.GET("/list", r.statefulSetController.GetStatefulSetList)
		statefulSetGroup.GET("/detail", r.statefulSetController.GetStatefulSetDetail)
		statefulSetGroup.POST("/create", r.statefulSetController.CreateStatefulSet)
		statefulSetGroup.PUT("/update", r.statefulSetController.UpdateStatefulSet)
		statefulSetGroup.DELETE("/delete", r.statefulSetController.DeleteStatefulSet)
		statefulSetGroup.PUT("/scale", r.statefulSetController.ScaleStatefulSet)
	}
}
