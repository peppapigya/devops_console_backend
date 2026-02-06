package helm

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"helm.sh/helm/v3/pkg/chart/loader"
)

// ChartService Chart相关服务
type ChartService struct{}

// NewChartService 创建Chart服务实例
func NewChartService() *ChartService {
	return &ChartService{}
}

// GetDefaultValues 获取 Chart 的默认 values.yaml
func (s *ChartService) GetDefaultValues(chartURL string) (string, error) {
	// 1. 下载 Chart 压缩包
	resp, err := http.Get(chartURL)
	if err != nil {
		return "", fmt.Errorf("下载Chart失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载Chart失败: HTTP %d", resp.StatusCode)
	}

	// 2. 读取响应体
	chartData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取Chart数据失败: %w", err)
	}

	// 3. 使用 Helm SDK 加载 Chart
	chart, err := loader.LoadArchive(bytes.NewReader(chartData))
	if err != nil {
		return "", fmt.Errorf("解析Chart失败: %w", err)
	}

	// 4. 返回 values.yaml 内容
	// 4. 从 Raw 文件中找到 values.yaml
	for _, f := range chart.Raw {
		if f.Name == "values.yaml" {
			return string(f.Data), nil
		}
	}

	return "", fmt.Errorf("未在 Chart 中找到 values.yaml")
}
