package helm

import (
	"encoding/json"
	"fmt"

	"devops-console-backend/internal/dal"

	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

// ReleaseService Helm Release管理服务
type ReleaseService struct {
	db *gorm.DB
}

// NewReleaseService 创建Release服务实例
func NewReleaseService(db *gorm.DB) *ReleaseService {
	return &ReleaseService{db: db}
}

// InstallRequest 安装请求参数
type InstallRequest struct {
	InstanceID   uint
	Namespace    string
	ReleaseName  string
	ChartName    string
	ChartVersion string
	RepoName     string
	Values       map[string]interface{}
}

// InstallChart 安装Helm Chart
// 注意: 此函数需要Helm SDK依赖，当前为示例实现
func (s *ReleaseService) InstallChart(req InstallRequest) error {

	// 1. 初始化ActionConfig
	actionConfig, err := NewActionConfig(req.Namespace, req.InstanceID)
	if err != nil {
		return fmt.Errorf("初始化 ActionConfig 失败 %w", err)
	}

	// 2. 创建Install Action
	client := action.NewInstall(actionConfig)
	client.Namespace = req.Namespace
	client.ReleaseName = req.ReleaseName
	client.Version = req.ChartVersion

	// 3. 加载Chart
	chartPath := fmt.Sprintf("%s/%s", req.RepoName, req.ChartName)
	chart, err := client.ChartPathOptions.LocateChart(chartPath, cli.New())
	if err != nil {
		return fmt.Errorf("failed to locate chart: %w", err)
	}

	chartRequested, err := loader.Load(chart)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}

	// 4. 执行安装
	release, err := client.Run(chartRequested, req.Values)
	if err != nil {
		return fmt.Errorf("failed to install chart: %w", err)
	}

	// 5. 记录到数据库
	valuesJSON, _ := json.Marshal(req.Values)
	helmRelease := dal.HelmRelease{
		InstanceID:   req.InstanceID,
		Namespace:    req.Namespace,
		ReleaseName:  req.ReleaseName,
		ChartName:    req.ChartName,
		ChartVersion: req.ChartVersion,
		Status:       release.Info.Status.String(),
		Values:       string(valuesJSON),
	}

	return s.db.Create(&helmRelease).Error
}

// UninstallRelease 卸载Helm Release
func (s *ReleaseService) UninstallRelease(instanceID uint, namespace, releaseName string) error {
	// 1. 初始化ActionConfig
	actionConfig, err := NewActionConfig(namespace, instanceID)
	if err != nil {
		return fmt.Errorf("无法创建action config: %w", err)
	}
	// 2. 创建Uninstall Action
	client := action.NewUninstall(actionConfig)
	// 3. 执行卸载
	_, err = client.Run(releaseName)
	if err != nil {
		return fmt.Errorf("无法卸载release: %w", err)
	}
	// 4. 更新数据库状态
	return s.db.Model(&dal.HelmRelease{}).
		Where("instance_id = ? AND namespace = ? AND release_name = ?", instanceID, namespace, releaseName).
		Update("status", "uninstalled").Error
}

// ListReleases 列出已安装的Release
func (s *ReleaseService) ListReleases(instanceID uint, namespace string) ([]dal.HelmRelease, error) {
	var releases []dal.HelmRelease
	query := s.db.Where("instance_id = ?", instanceID)

	if namespace != "" && namespace != "all" {
		query = query.Where("namespace = ?", namespace)
	}

	err := query.Find(&releases).Error
	return releases, err
}

// GetReleaseDetail 获取Release详情
func (s *ReleaseService) GetReleaseDetail(instanceID uint, namespace, releaseName string) (*dal.HelmRelease, error) {
	var release dal.HelmRelease
	err := s.db.Where("instance_id = ? AND namespace = ? AND release_name = ?",
		instanceID, namespace, releaseName).First(&release).Error

	if err != nil {
		return nil, err
	}

	// TODO: @dxg 可以从K8s获取实时状态
	actionConfig, _ := NewActionConfig(namespace, instanceID)
	getClient := action.NewGet(actionConfig)
	rel, _ := getClient.Run(releaseName)
	release.Status = rel.Info.Status.String()

	return &release, nil
}
