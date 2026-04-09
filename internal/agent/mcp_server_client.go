package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ToolCallRequest 发送给 mcp-agent 工具服务器的请求
type ToolCallRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Command  string `json:"command"`
	Cwd      string `json:"cwd"`
	Timeout  int    `json:"timeout"`
}

// ToolCallResult mcp-agent 工具服务器返回的结果
type ToolCallResult struct {
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// mcpToolResponse mcp-agent HTTP 响应包装
type mcpToolResponse struct {
	Data    *ToolCallResult `json:"data"`
	Message string          `json:"message"`
}

// MCPServerClient 向 mcp-agent 工具服务器发送工具调用请求
type MCPServerClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewMCPServerClient(baseURL, token string) *MCPServerClient {
	return &MCPServerClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // SSH 命令可能耗时较长
		},
	}
}

// ExecuteSSH 调用 mcp-agent 的 /api/v1/tools/execute_ssh 接口
func (c *MCPServerClient) ExecuteSSH(req ToolCallRequest) (*ToolCallResult, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	url := c.baseURL + "/api/v1/tools/execute_ssh"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("构建 HTTP 请求失败: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("token", c.token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mcp-agent 返回异常状态码: %d", resp.StatusCode)
	}

	var result mcpToolResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	if result.Data == nil {
		return nil, fmt.Errorf("mcp-agent 返回空结果")
	}
	return result.Data, nil
}
