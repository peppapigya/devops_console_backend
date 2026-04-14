package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ToolCallRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Command  string `json:"command"`
	Cwd      string `json:"cwd"`
	Timeout  int    `json:"timeout"`
}

type StructuredToolBase struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Timeout  int    `json:"timeout"`
}

type ServiceStatusRequest struct {
	StructuredToolBase
	Service string `json:"service"`
}

type ServiceLogsRequest struct {
	StructuredToolBase
	Service string `json:"service"`
	Lines   int    `json:"lines"`
}

type FileSnippetRequest struct {
	StructuredToolBase
	FilePath  string `json:"file_path"`
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
}

type NginxConfigValidateRequest struct {
	StructuredToolBase
	ConfigPath string `json:"config_path"`
}

type FileReplaceRequest struct {
	StructuredToolBase
	FilePath     string `json:"file_path"`
	Search       string `json:"search"`
	Replace      string `json:"replace"`
	CreateBackup bool   `json:"create_backup"`
}

type RestartServiceRequest struct {
	StructuredToolBase
	Service string `json:"service"`
}

type ToolCallResult struct {
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type mcpToolResponse struct {
	Data    *ToolCallResult `json:"data"`
	Message string          `json:"message"`
}

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
			Timeout: 5 * time.Minute,
		},
	}
}

func (c *MCPServerClient) ExecuteSSH(req ToolCallRequest) (*ToolCallResult, error) {
	return c.post("/api/v1/tools/execute_ssh", req)
}

func (c *MCPServerClient) InspectServiceStatus(req ServiceStatusRequest) (*ToolCallResult, error) {
	return c.post("/api/v1/tools/inspect_service_status", req)
}

func (c *MCPServerClient) InspectServiceLogs(req ServiceLogsRequest) (*ToolCallResult, error) {
	return c.post("/api/v1/tools/inspect_service_logs", req)
}

func (c *MCPServerClient) InspectFileSnippet(req FileSnippetRequest) (*ToolCallResult, error) {
	return c.post("/api/v1/tools/inspect_file_snippet", req)
}

func (c *MCPServerClient) ValidateNginxConfig(req NginxConfigValidateRequest) (*ToolCallResult, error) {
	return c.post("/api/v1/tools/validate_nginx_config", req)
}

func (c *MCPServerClient) ReplaceFileContent(req FileReplaceRequest) (*ToolCallResult, error) {
	return c.post("/api/v1/tools/replace_file_content", req)
}

func (c *MCPServerClient) RestartService(req RestartServiceRequest) (*ToolCallResult, error) {
	return c.post("/api/v1/tools/restart_service", req)
}

func (c *MCPServerClient) post(path string, payload interface{}) (*ToolCallResult, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("build request failed: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("token", c.token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mcp server returned status %d", resp.StatusCode)
	}

	var result mcpToolResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %v", err)
	}
	if result.Data == nil {
		return nil, fmt.Errorf("mcp server returned empty data")
	}
	return result.Data, nil
}
