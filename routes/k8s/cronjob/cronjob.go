package cronjob

import (
	"devops-console-backend/controllers/k8s/cronjob"
	"github.com/gin-gonic/gin"
)

// CronJobRoute CronJob路由
type CronJobRoute struct {
	controller *cronjob.CronJobController
}

// NewCronJobRoute 创建CronJob路由实例
func NewCronJobRoute() *CronJobRoute {
	return &CronJobRoute{
		controller: cronjob.NewCronJobController(),
	}
}

// RegisterSubRouter 注册子路由
func (r *CronJobRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	cronjobGroup := apiGroup.Group("/k8s/cronjob")
	{
		cronjobGroup.POST("/create", r.controller.CreateCronJob)
		cronjobGroup.DELETE("/delete", r.controller.DeleteCronJob)
		cronjobGroup.GET("/list/:namespace", r.controller.GetCronJobList)
		cronjobGroup.PUT("/update", r.controller.UpdateCronJob)
	}
}
