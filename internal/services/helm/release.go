package helm

import (
	"fmt"

	"devops-console-backend/internal/dal"

	"sigs.k8s.io/yaml"

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
	ChartURL     string
	Values       string
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
	client.CreateNamespace = true // 自动创建命名空间

	// 3. 加载Chart
	var chartPath string
	if req.ChartURL != "" {
		chartPath = req.ChartURL
	} else {
		chartPath = fmt.Sprintf("%s/%s", req.RepoName, req.ChartName)
	}

	chart, err := client.ChartPathOptions.LocateChart(chartPath, cli.New())
	if err != nil {
		return fmt.Errorf("failed to locate chart: %w", err)
	}

	chartRequested, err := loader.Load(chart)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}

	// 4. 解析Values YAML 到 map
	vals := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(req.Values), &vals); err != nil {
		return fmt.Errorf("failed to parse values yaml: %w", err)
	}

	// 5. 执行安装
	release, err := client.Run(chartRequested, vals)
	if err != nil {
		return fmt.Errorf("failed to install chart: %w", err)
	}

	// 6. 记录到数据库 (直接存储原始YAML字符串)
	helmRelease := dal.HelmRelease{
		InstanceID:   req.InstanceID,
		Namespace:    req.Namespace,
		ReleaseName:  req.ReleaseName,
		ChartName:    req.ChartName,
		ChartVersion: req.ChartVersion,
		Status:       release.Info.Status.String(),
		Values:       req.Values, // 存储原始YAML
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
	// 4. 从数据库删除记录
	return s.db.Where("instance_id = ? AND namespace = ? AND release_name = ?", instanceID, namespace, releaseName).
		Delete(&dal.HelmRelease{}).Error
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

// UpgradeRequest 升级请求参数
type UpgradeRequest struct {
	InstanceID   uint
	Namespace    string
	ReleaseName  string
	ChartName    string
	ChartVersion string
	RepoName     string
	RepoURL      string
	ChartURL     string
	Values       string
}

// UpgradeRelease 升级Helm Release
func (s *ReleaseService) UpgradeRelease(req UpgradeRequest) error {
	// 1. 初始化ActionConfig
	actionConfig, err := NewActionConfig(req.Namespace, req.InstanceID)
	if err != nil {
		return fmt.Errorf("初始化 ActionConfig 失败 %w", err)
	}

	// 2. 创建Upgrade Action
	client := action.NewUpgrade(actionConfig)
	client.Namespace = req.Namespace
	client.Version = req.ChartVersion

	// 2.5 设置 RepoURL
	if req.RepoURL != "" {
		client.ChartPathOptions.RepoURL = req.RepoURL
	}

	// 3. 加载Chart
	chartPath := req.ChartName

	// 如果提供了 ChartURL，直接使用 (通常用于直接指定tgz地址)
	if req.ChartURL != "" {
		chartPath = req.ChartURL
	} else if req.RepoName != "" {
		// 如果有 RepoName 且没有 RepoURL ，尝试使用 RepoName/ChartName 格式
		// 但如果有 RepoURL，LocateChart 会使用 RepoURL + ChartName
		if req.RepoURL == "" {
			chartPath = fmt.Sprintf("%s/%s", req.RepoName, req.ChartName)
		}
	}

	cp, err := client.ChartPathOptions.LocateChart(chartPath, cli.New())
	if err != nil {
		return fmt.Errorf("failed to locate chart: %w", err)
	}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}

	// 4. 解析Values YAML 到 map
	vals := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(req.Values), &vals); err != nil {
		return fmt.Errorf("failed to parse values yaml: %w", err)
	}

	// 5. 执行升级
	release, err := client.Run(req.ReleaseName, chartRequested, vals)
	if err != nil {
		return fmt.Errorf("failed to upgrade release: %w", err)
	}

	// 6. 更新数据库记录
	return s.db.Model(&dal.HelmRelease{}).
		Where("instance_id = ? AND namespace = ? AND release_name = ?", req.InstanceID, req.Namespace, req.ReleaseName).
		Updates(map[string]interface{}{
			"chart_version": req.ChartVersion,
			"status":        release.Info.Status.String(),
			"values":        req.Values,
			"updated_at":    gorm.Expr("NOW()"),
		}).Error
}
