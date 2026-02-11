package helm

import (
	helmController "devops-console-backend/internal/controllers/helm"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HelmRoute Helm路由
type HelmRoute struct {
	repoController    *helmController.RepoController
	chartController   *helmController.ChartController
	releaseController *helmController.ReleaseController
}

// NewHelmRoute 创建Helm路由实例
func NewHelmRoute(db *gorm.DB) *HelmRoute {
	return &HelmRoute{
		repoController:    helmController.NewRepoController(db),
		chartController:   helmController.NewChartController(db),
		releaseController: helmController.NewReleaseController(db),
	}
}

// RegisterSubRouter 注册Helm子路由
func (r *HelmRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	// Helm 仓库管理路由
	repoGroup := apiGroup.Group("/helm/repos")
	{
		repoGroup.GET("", r.repoController.GetRepoList)
		repoGroup.POST("", r.repoController.CreateRepo)
		repoGroup.PUT("/:id", r.repoController.UpdateRepo)
		repoGroup.DELETE("/:id", r.repoController.DeleteRepo)
		repoGroup.POST("/:id/sync", r.repoController.SyncRepo)
	}

	// Helm Chart管理路由
	chartGroup := apiGroup.Group("/helm/charts")
	{
		chartGroup.GET("", r.chartController.GetChartList)
		chartGroup.GET("/:name/versions", r.chartController.GetChartVersions)
		chartGroup.GET("/:name/values", r.chartController.GetDefaultValues)
	}

	// Helm Release管理路由
	releaseGroup := apiGroup.Group("/helm/releases")
	{
		releaseGroup.GET("", r.releaseController.GetReleaseList)
		releaseGroup.GET("/:namespace/:name", r.releaseController.GetReleaseDetail)
		releaseGroup.DELETE("/:namespace/:name", r.releaseController.UninstallRelease)
	}

	// Helm安装/升级路由
	apiGroup.POST("/helm/install", r.releaseController.InstallChart)
	apiGroup.PUT("/helm/upgrade", r.releaseController.UpgradeRelease)
}
