package cicd

import (
	"devops-console-backend/cmd/generate/wireInfo"

	"github.com/gin-gonic/gin"
)

func RegisterCiCdRouters(router *gin.RouterGroup) {
	pipelineController := wireInfo.InitializePipelineController()
	pipelineGroup := router.Group("/pipelines")
	{
		pipelineGroup.GET("/page", pipelineController.GetPagePipelines)
		pipelineGroup.GET("/:id", pipelineController.GetPipelineById)
		pipelineGroup.DELETE("/:id", pipelineController.DeletePipeline)
		pipelineGroup.PUT("/:id", pipelineController.UpdatePipeline)
		pipelineGroup.POST("/", pipelineController.CreatePipeline)
	}

	pipelineRunController := wireInfo.InitializePipelineRunsController()
	pipelineRunRouter := router.Group("/pipeline-runs")
	{
		pipelineRunRouter.GET("/:id", pipelineRunController.GetPipelineRunById)
		pipelineRunRouter.DELETE("/:id", pipelineRunController.DeletePipelineRun)
		pipelineRunRouter.PUT("/", pipelineRunController.UpdatePipelineRun)
		pipelineRunRouter.GET("/page", pipelineRunController.GetPagePipelineRuns)
	}

	projectsMapper := wireInfo.InitializeProjectsController()
	projectsRouter := router.Group("/projects")
	{
		projectsRouter.GET("/:id", projectsMapper.GetProjectById)
		projectsRouter.DELETE("/:id", projectsMapper.DeleteProject)
		projectsRouter.PUT("/", projectsMapper.UpdateProject)
		projectsRouter.GET("/page", projectsMapper.GetPageProjects)
		projectsRouter.GET("/list", projectsMapper.GetProjects)
	}
	argoController := wireInfo.InitializeArgoController()
	argoRouter := router.Group("/argo")
	{
		argoRouter.POST("/execute", argoController.ExecutePipeline)
	}

	pipelineStepsController := wireInfo.InitializePipelineStepsController()
	pipelineStepsRouter := router.Group("/pipeline-steps")
	{
		pipelineStepsRouter.GET("/list", pipelineStepsController.GetPipelineSteps)
		pipelineStepsRouter.GET("/:id", pipelineStepsController.GetPipelineStepById)
		pipelineStepsRouter.POST("/", pipelineStepsController.CreatePipelineStep)
		pipelineStepsRouter.PUT("/", pipelineStepsController.UpdatePipelineStep)
		pipelineStepsRouter.DELETE("/:id", pipelineStepsController.DeletePipelineStep)
	}
}
