package helm

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"devops-console-backend/internal/dal"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// RepoService Helm仓库服务
type RepoService struct {
	db *gorm.DB
}

// NewRepoService 创建仓库服务实例
func NewRepoService(db *gorm.DB) *RepoService {
	return &RepoService{db: db}
}

// IndexFile Helm仓库index.yaml结构
type IndexFile struct {
	APIVersion string                    `yaml:"apiVersion"`
	Entries    map[string][]ChartVersion `yaml:"entries"`
	Generated  time.Time                 `yaml:"generated"`
}

// ChartVersion Chart版本信息
type ChartVersion struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	AppVersion  string   `yaml:"appVersion"`
	Description string   `yaml:"description"`
	Icon        string   `yaml:"icon"`
	URLs        []string `yaml:"urls"`
}

// SyncRepo 同步仓库：下载index.yaml并更新数据库
func (s *RepoService) SyncRepo(repoID uint) error {
	// 1. 获取仓库信息
	var repo dal.HelmRepo
	if err := s.db.First(&repo, repoID).Error; err != nil {
		return fmt.Errorf("无法找到对应的仓库: %w", err)
	}

	// 2. 下载index.yaml
	indexURL := repo.URL + "/index.yaml"
	resp, err := http.Get(indexURL)
	if err != nil {
		return fmt.Errorf("无法下载 index.yaml: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("无法下载 index.yaml: 错误码 %d", resp.StatusCode)
	}

	// 3. 解析index.yaml
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("无法读取yaml文件: %w", err)
	}

	var index IndexFile
	if err := yaml.Unmarshal(body, &index); err != nil {
		return fmt.Errorf("无法解析YAML文件: %w", err)
	}

	// 4. 批量更新HelmChart表
	// 分批删除旧数据，避免大事务锁超时
	for {
		result := s.db.Where("repo_id = ?", repoID).Limit(500).Delete(&dal.HelmChart{})
		if result.Error != nil {
			return fmt.Errorf("删除旧Chart数据失败: %w", result.Error)
		}
		// 如果没有删除任何记录，说明已经删完了
		if result.RowsAffected == 0 {
			break
		}
	}

	// 收集所有 Chart 数据
	charts := make([]dal.HelmChart, 0, 1000)
	for chartName, versions := range index.Entries {
		for _, version := range versions {
			chart := dal.HelmChart{
				RepoID:      repoID,
				Name:        chartName,
				Version:     version.Version,
				AppVersion:  version.AppVersion,
				Description: version.Description,
				Icon:        version.Icon,
				ChartURL:    getChartURL(version.URLs, repo.URL),
			}
			charts = append(charts, chart)
		}
	}

	// 分批插入，每批 200 条，避免单个事务过大导致锁超时
	batchSize := 200
	for i := 0; i < len(charts); i += batchSize {
		end := i + batchSize
		if end > len(charts) {
			end = len(charts)
		}
		batch := charts[i:end]

		if err := s.db.CreateInBatches(batch, batchSize).Error; err != nil {
			return fmt.Errorf("批量插入Chart数据失败: %w", err)
		}
	}

	return nil
}

// getChartURL 从URLs列表中获取完整的Chart下载地址
func getChartURL(urls []string, repoURL string) string {
	if len(urls) == 0 {
		return ""
	}
	// 如果URL是相对路径，拼接仓库地址
	url := urls[0]
	if len(url) > 0 && url[0] != 'h' {
		return repoURL + "/" + url
	}
	return url
}
