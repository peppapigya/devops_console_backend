package helm

import (
	"devops-console-backend/internal/dal/request/helm"
	helmService "devops-console-backend/internal/services/helm"
	"devops-console-backend/pkg/utils"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReleaseController Helm Release控制器
type ReleaseController struct {
	db             *gorm.DB
	releaseService *helmService.ReleaseService
}

// NewReleaseController 创建Release控制器实例
func NewReleaseController(db *gorm.DB) *ReleaseController {
	return &ReleaseController{
		db:             db,
		releaseService: helmService.NewReleaseService(db),
	}
}

// InstallChart 安装Helm Chart
func (c *ReleaseController) InstallChart(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	var req helm.InstallChartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helper.ValidationError(err.Error())
		return
	}

	// 调用服务层安装
	installReq := helmService.InstallRequest{
		InstanceID:   req.InstanceID,
		Namespace:    req.Namespace,
		ReleaseName:  req.ReleaseName,
		ChartName:    req.ChartName,
		ChartVersion: req.ChartVersion,
		RepoName:     req.RepoName,
		Values:       req.Values,
	}

	if err := c.releaseService.InstallChart(installReq); err != nil {
		helper.InternalError(fmt.Sprintf("安装失败: %s", err.Error()))
		return
	}

	helper.Success("安装成功")
}

// UninstallRelease 卸载Helm Release
func (c *ReleaseController) UninstallRelease(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	var instanceID uint
	namespace := ctx.Param("namespace")
	releaseName := ctx.Param("name")
	utils.GetParam(ctx, "instance_id", &instanceID, nil)

	if err := c.releaseService.UninstallRelease(instanceID, namespace, releaseName); err != nil {
		helper.InternalError(fmt.Sprintf("卸载失败: %s", err.Error()))
		return
	}

	helper.Success("卸载成功")
}

// GetReleaseList 获取已安装的Release列表
func (c *ReleaseController) GetReleaseList(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	var instanceID uint
	namespace := ctx.Query("namespace")
	utils.GetParam(ctx, "instance_id", &instanceID, nil)

	releases, err := c.releaseService.ListReleases(instanceID, namespace)
	if err != nil {
		helper.DatabaseError("获取 Release 列表失败")
		return
	}

	// 转换为响应格式
	releaseList := make([]helm.ReleaseListItem, 0, len(releases))
	for _, rel := range releases {
		releaseList = append(releaseList, helm.ReleaseListItem{
			ID:           rel.ID,
			ReleaseName:  rel.ReleaseName,
			Namespace:    rel.Namespace,
			ChartName:    rel.ChartName,
			ChartVersion: rel.ChartVersion,
			Status:       rel.Status,
			UpdatedAt:    rel.UpdatedAt.Unix(),
		})
	}

	helper.SuccessWithData("获取成功", "releaseList", releaseList)
}

// GetReleaseDetail 获取Release详情
func (c *ReleaseController) GetReleaseDetail(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	var instanceID uint
	namespace := ctx.Param("namespace")
	releaseName := ctx.Param("name")
	utils.GetParam(ctx, "instance_id", &instanceID, nil)

	release, err := c.releaseService.GetReleaseDetail(instanceID, namespace, releaseName)
	if err != nil {
		helper.DatabaseError("获取 Release 详情失败")
		return
	}

	helper.SuccessWithData("获取成功", "release", release)
}
