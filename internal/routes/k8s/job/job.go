package job

import (
	"devops-console-backend/internal/controllers/k8s/job"

	"github.com/gin-gonic/gin"
)

// JobRoute Job路由
type JobRoute struct {
	controller *job.JobController
}

// NewJobRoute 创建Job路由实例
func NewJobRoute() *JobRoute {
	return &JobRoute{
		controller: job.NewJobController(),
	}
}

// RegisterSubRouter 注册子路由
func (r *JobRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	jobGroup := apiGroup.Group("/k8s/job")
	{
		jobGroup.GET("/detail/:namespace/:jobName", r.controller.GetJobDetail)
		jobGroup.GET("/list/:namespace", r.controller.GetJobList)
		jobGroup.POST("/create", r.controller.CreateJob)
		jobGroup.DELETE("/delete/:namespace/:jobName", r.controller.DeleteJob)
	}
}
