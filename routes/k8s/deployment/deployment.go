package deployment

import (
	"devops-console-backend/controllers/k8s/deployment"
	"github.com/gin-gonic/gin"
)

// DeploymentRoute Deployment路由
type DeploymentRoute struct {
	controller *deployment.DeploymentController
}

// NewDeploymentRoute 创建Deployment路由实例
func NewDeploymentRoute() *DeploymentRoute {
	return &DeploymentRoute{
		controller: deployment.NewDeploymentController(),
	}
}

// RegisterSubRouter 注册子路由
func (r *DeploymentRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	deploymentGroup := apiGroup.Group("/k8s/deployment")
	{
		deploymentGroup.GET("/detail/:namespace/:deploymentName", r.controller.GetDeploymentDetail)
		deploymentGroup.GET("/list/:namespace", r.controller.GetDeploymentList)
		deploymentGroup.POST("/create/:namespace", r.controller.CreateDeployment)
		deploymentGroup.PUT("/update/:namespace/:deploymentName", r.controller.UpdateDeployment)
		deploymentGroup.PUT("/scale/:namespace/:deploymentName/:replicas", r.controller.ScaleDeployment)
		deploymentGroup.DELETE("/delete/:namespace/:deploymentName", r.controller.DeleteDeployment)
	}
}