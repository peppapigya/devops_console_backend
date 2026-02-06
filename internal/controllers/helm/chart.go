package helm

import (
	"devops-console-backend/internal/dal"
	"devops-console-backend/internal/dal/request/helm"
	helmService "devops-console-backend/internal/services/helm"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ChartController Helm Chart控制器
type ChartController struct {
	db *gorm.DB
}

// NewChartController 创建Chart控制器实例
func NewChartController(db *gorm.DB) *ChartController {
	return &ChartController{db: db}
}

// GetChartList 获取Chart列表（支持搜索和分页）
func (c *ChartController) GetChartList(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	var req helm.ChartListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		helper.ValidationError(err.Error())
		return
	}

	// 构建基础查询条件
	baseWhere := ""
	baseArgs := []interface{}{}

	if req.RepoID > 0 {
		baseWhere += " AND hc.repo_id = ?"
		baseArgs = append(baseArgs, req.RepoID)
	}

	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		baseWhere += " AND (hc.name LIKE ? OR hc.description LIKE ?)"
		baseArgs = append(baseArgs, keyword, keyword)
	}

	// 使用子查询获取每个 chart 名称的最新记录（按 created_at 排序）
	// 使用 ROW_NUMBER 窗口函数来为每个名称分组并排序
	subQuery := `
		SELECT hc.*, hr.name as repo_name,
		       ROW_NUMBER() OVER (PARTITION BY hc.name ORDER BY hc.created_at DESC) as rn
		FROM helm_chart hc
		LEFT JOIN helm_repo hr ON hc.repo_id = hr.id
		WHERE 1=1` + baseWhere

	// 计算总数（去重后的chart名称数量）
	var total int64
	countQuery := `
		SELECT COUNT(DISTINCT name) 
		FROM helm_chart hc 
		WHERE 1=1` + baseWhere

	if err := c.db.Raw(countQuery, baseArgs...).Count(&total).Error; err != nil {
		helper.DatabaseError("获取总数失败")
		return
	}

	// 主查询：获取分页数据
	// 从子查询结果中只选择 rn=1 的记录（每个名称的最新版本）
	type ChartWithRepo struct {
		dal.HelmChart
		RepoName string `gorm:"column:repo_name"`
	}

	var charts []ChartWithRepo
	mainQuery := `
		SELECT * FROM (` + subQuery + `) as ranked
		WHERE rn = 1
		ORDER BY name
		LIMIT ? OFFSET ?`

	offset := (req.Page - 1) * req.PageSize
	queryArgs := append(baseArgs, req.PageSize, offset)

	if err := c.db.Raw(mainQuery, queryArgs...).Scan(&charts).Error; err != nil {
		helper.DatabaseError("获取 Chart 列表失败")
		return
	}

	// 转换为响应格式
	chartList := make([]helm.ChartListItem, 0, len(charts))
	for _, chart := range charts {
		chartList = append(chartList, helm.ChartListItem{
			ID:          chart.ID,
			RepoID:      chart.RepoID,
			RepoName:    chart.RepoName,
			Name:        chart.Name,
			Version:     chart.Version,
			AppVersion:  chart.AppVersion,
			Description: chart.Description,
			Icon:        chart.Icon,
			ChartURL:    chart.ChartURL,
		})
	}

	helper.SuccessWithData("获取成功", "data", gin.H{
		"chartList": chartList,
		"total":     total,
		"page":      req.Page,
		"pageSize":  req.PageSize,
	})
}

// GetChartVersions 获取Chart的所有版本
func (c *ChartController) GetChartVersions(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)

	chartName := ctx.Param("name")
	var repoID uint
	utils.GetParam(ctx, "repo_id", &repoID, nil)

	var versions []dal.HelmChart
	query := c.db.Where("name = ?", chartName)
	if repoID > 0 {
		query = query.Where("repo_id = ?", repoID)
	}

	if err := query.Order("created_at DESC").Find(&versions).Error; err != nil {
		helper.DatabaseError("获取 Chart版本列表失败")
		return
	}

	versionList := make([]helm.ChartVersionListItem, 0, len(versions))
	for _, v := range versions {
		versionList = append(versionList, helm.ChartVersionListItem{
			Version:    v.Version,
			AppVersion: v.AppVersion,
			CreatedAt:  v.CreatedAt.Unix(),
		})
	}

	helper.SuccessWithData("获取成功", "versions", versionList)
}

// GetDefaultValues 获取Chart的默认values.yaml
func (c *ChartController) GetDefaultValues(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	chartName := ctx.Param("name")
	var repoID uint
	var version string
	utils.GetParam(ctx, "repo_id", &repoID, nil)
	utils.GetParam(ctx, "version", &version, nil)
	// 查找 Chart记录
	var chart dal.HelmChart
	query := c.db.Where("name = ?", chartName)
	if repoID > 0 {
		query = query.Where("repo_id = ?", repoID)
	}
	if version != "" {
		query = query.Where("version = ?", version)
	} else {
		query = query.Order("created_at DESC")
	}
	if err := query.First(&chart).Error; err != nil {
		helper.NotFound("Chart 不存在")
		return
	}
	// 下载并解析 Chart
	chartService := helmService.NewChartService()
	values, err := chartService.GetDefaultValues(chart.ChartURL)
	if err != nil {
		helper.Error(500, "获取默认配置失败: "+err.Error())
		return
	}
	helper.SuccessWithData("获取成功", "values", values)
}
