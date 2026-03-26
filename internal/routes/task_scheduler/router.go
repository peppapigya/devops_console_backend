package task_scheduler

import (
	tsCtrl "devops-console-backend/internal/controllers/task_scheduler"
	"devops-console-backend/internal/dal/mapper"
	tsSvc "devops-console-backend/internal/services/task_scheduler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterTaskSchedulerRouters(apiGroup *gin.RouterGroup, db *gorm.DB) {
	workflowMapper := mapper.NewTaskWorkflowMapper(db)
	nodeMapper := mapper.NewTaskNodeMapper(db)
	edgeMapper := mapper.NewTaskEdgeMapper(db)
	executionMapper := mapper.NewTaskExecutionMapper(db)
	nodeExecMapper := mapper.NewTaskNodeExecutionMapper(db)

	workflowService := tsSvc.NewTaskWorkflowService(workflowMapper, nodeMapper, edgeMapper, executionMapper, nodeExecMapper)

	workflowCtrl := tsCtrl.NewTaskWorkflowController(workflowService)
	executionCtrl := tsCtrl.NewTaskExecutionController(executionMapper, nodeExecMapper)

	tsGroup := apiGroup.Group("/task-scheduler")
	{
		workflowGroup := tsGroup.Group("/workflows")
		{
			workflowGroup.GET("", workflowCtrl.List)
			workflowGroup.GET("/:id", workflowCtrl.GetByID)
			workflowGroup.POST("", workflowCtrl.Create)
			workflowGroup.PUT("/:id", workflowCtrl.Update)
			workflowGroup.DELETE("/:id", workflowCtrl.Delete)
			workflowGroup.PUT("/:id/status", workflowCtrl.UpdateStatus)
			workflowGroup.POST("/:id/execute", workflowCtrl.Execute)
		}
		tsGroup.GET("/executions", executionCtrl.ListExecutions)
		tsGroup.GET("/executions/:id", executionCtrl.GetExecutionDetail)
		tsGroup.GET("/executions/:id/logs", executionCtrl.GetExecutionLogs)
	}
}
