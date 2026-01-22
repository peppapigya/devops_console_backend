package service

import (
	"devops-console-backend/controllers/k8s/service"
	"github.com/gin-gonic/gin"
)

// ServiceRoute Service路由
type ServiceRoute struct {
	controller *service.ServiceController
}

// NewServiceRoute 创建Service路由实例
func NewServiceRoute() *ServiceRoute {
	return &ServiceRoute{
		controller: service.NewServiceController(),
	}
}

// RegisterSubRouter 注册子路由
func (r *ServiceRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	serviceGroup := apiGroup.Group("/k8s/service")
	{
		serviceGroup.GET("/detail/:namespace/:name", r.controller.GetServiceDetail)
		serviceGroup.GET("/list/:namespace", r.controller.GetServiceList)
		serviceGroup.POST("/create/:namespace/:name", r.controller.CreateService)
		serviceGroup.PUT("/update/:namespace/:name", r.controller.UpdateService)
		serviceGroup.DELETE("/delete/:namespace/:name", r.controller.DeleteService)
		serviceGroup.DELETE("/delete-multiple/:namespace", r.controller.DeleteMultipleServices)
		serviceGroup.DELETE("/delete-by-label/:namespace", r.controller.DeleteServiceByLabel)
	}
}