package helm

// HelmRepo Request/Response Types

// RepoCreateRequest 创建仓库请求
type RepoCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	URL      string `json:"url" binding:"required,url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// RepoUpdateRequest 更新仓库请求
type RepoUpdateRequest struct {
	Name     string `json:"name"`
	URL      string `json:"url" binding:"omitempty,url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// RepoListItem 仓库列表项
type RepoListItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Username  string `json:"username,omitempty"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// HelmChart Request/Response Types

// ChartListRequest Chart列表请求
type ChartListRequest struct {
	RepoID   uint   `form:"repo_id"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
}

// ChartListItem Chart列表项
type ChartListItem struct {
	ID          uint   `json:"id"`
	RepoID      uint   `json:"repo_id"`
	RepoName    string `json:"repo_name"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	ChartURL    string `json:"chart_url"`
}

// ChartVersionListItem Chart版本列表项
type ChartVersionListItem struct {
	Version    string `json:"version"`
	AppVersion string `json:"app_version"`
	CreatedAt  int64  `json:"created_at"`
	RepoURL    string `json:"repo_url"`
}

// HelmRelease Request/Response Types

// InstallChartRequest 安装Chart请求
type InstallChartRequest struct {
	InstanceID   uint   `json:"instance_id" binding:"required"`
	Namespace    string `json:"namespace" binding:"required"`
	ReleaseName  string `json:"release_name" binding:"required"`
	ChartName    string `json:"chart_name" binding:"required"`
	ChartVersion string `json:"chart_version"`
	RepoName     string `json:"repo_name" binding:"required"`
	ChartURL     string `json:"chart_url"`
	Values       string `json:"values"` // 自定义values (YAML字符串)
}

// UpgradeChartRequest 升级Chart请求
type UpgradeChartRequest struct {
	InstanceID   uint   `json:"instance_id" binding:"required"`
	Namespace    string `json:"namespace" binding:"required"`
	ReleaseName  string `json:"release_name" binding:"required"`
	ChartName    string `json:"chart_name"`
	ChartVersion string `json:"chart_version"`
	RepoName     string `json:"repo_name"`
	RepoURL      string `json:"repo_url"`
	ChartURL     string `json:"chart_url"`
	Values       string `json:"values"` // YAML字符串
}

// ReleaseListItem Release列表项
type ReleaseListItem struct {
	ID           uint   `json:"id"`
	ReleaseName  string `json:"release_name"`
	Namespace    string `json:"namespace"`
	ChartName    string `json:"chart_name"`
	ChartVersion string `json:"chart_version"`
	Status       string `json:"status"`
	Revision     int    `json:"revision"`
	UpdatedAt    int64  `json:"updated_at"`
}

// ReleaseDetail Release详情
type ReleaseDetail struct {
	ID           uint   `json:"id"`
	ReleaseName  string `json:"release_name"`
	Namespace    string `json:"namespace"`
	ChartName    string `json:"chart_name"`
	ChartVersion string `json:"chart_version"`
	Status       string `json:"status"`
	Revision     int    `json:"revision"`
	Values       string `json:"values"` // YAML字符串
	Manifest     string `json:"manifest"`
	Notes        string `json:"notes"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}
