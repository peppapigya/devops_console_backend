package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPExecutor HTTP请求执行器
type HTTPExecutor struct {
	client *http.Client
}

func NewHTTPExecutor() *HTTPExecutor {
	return &HTTPExecutor{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (e *HTTPExecutor) GetType() string {
	return "http"
}

func (e *HTTPExecutor) Validate(config map[string]interface{}) error {
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("URL不能为空")
	}
	return nil
}

func (e *HTTPExecutor) Execute(ctx context.Context, execCtx *TaskExecutionContext) *ExecutionResult {
	startTime := time.Now()

	// 解析配置
	url := getString(execCtx.Config, "url", "")
	method := getString(execCtx.Config, "method", "GET")
	headers := getMap(execCtx.Config, "headers")
	body := getString(execCtx.Config, "body", "")

	execCtx.Logger.Log("info", fmt.Sprintf("开始执行HTTP请求: %s %s", method, url))

	if len(headers) > 0 {
		maskedHeaders := make(map[string]string)
		for k, v := range headers {
			val := fmt.Sprintf("%v", v)
			if isSensitiveHeader(k) {
				val = "******"
			}
			maskedHeaders[k] = val
		}
		execCtx.Logger.Log("info", fmt.Sprintf("请求头: %v", maskedHeaders))
	}

	if body != "" {
		displayBody := body
		if len(displayBody) > 1000 {
			displayBody = displayBody[:1000] + "... (已截断)"
		}
		execCtx.Logger.Log("info", fmt.Sprintf("请求体: \n%s", displayBody))
	}

	// 创建请求
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		execCtx.Logger.Log("error", fmt.Sprintf("创建请求失败: %v", err))
		return &ExecutionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("创建请求失败: %v", err),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, fmt.Sprintf("%v", v))
	}

	// 发送请求
	resp, err := e.client.Do(req)
	if err != nil {
		execCtx.Logger.Log("error", fmt.Sprintf("请求失败: %v", err))
		return &ExecutionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("请求失败: %v", err),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		execCtx.Logger.Log("error", fmt.Sprintf("读取响应失败: %v", err))
		return &ExecutionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("读取响应失败: %v", err),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}

	duration := time.Since(startTime).Milliseconds()

	execCtx.Logger.Log("info", fmt.Sprintf("HTTP请求完成，状态码: %d, 耗时: %dms", resp.StatusCode, duration))

	maskedRespHeaders := make(map[string]string)
	for k, v := range resp.Header {
		val := strings.Join(v, ", ")
		if isSensitiveHeader(k) {
			val = "******"
		}
		maskedRespHeaders[k] = val
	}
	execCtx.Logger.Log("info", fmt.Sprintf("响应头: %v", maskedRespHeaders))

	if len(respBody) > 0 {
		displayRespBody := string(respBody)
		if len(displayRespBody) > 1000 {
			displayRespBody = displayRespBody[:1000] + "... (已截断)"
		}
		execCtx.Logger.Log("info", fmt.Sprintf("响应内容: \n%s", displayRespBody))
	}

	// 判断成功
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return &ExecutionResult{
		Success:  success,
		Output:   string(respBody),
		Duration: duration,
		OutputVars: map[string]interface{}{
			"status_code": resp.StatusCode,
			"headers":     resp.Header,
		},
	}
}

func isSensitiveHeader(name string) bool {
	name = strings.ToLower(name)
	sensitive := []string{"authorization", "cookie", "set-cookie", "token", "x-auth-token", "password", "secret", "key"}
	for _, s := range sensitive {
		if strings.Contains(name, s) {
			return true
		}
	}
	return false
}
