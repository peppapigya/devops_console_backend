package mcp

import (
	"bytes"
	"devops-console-backend/internal/types"
	"encoding/json"
	"fmt"
	"net/http"
)

type MCPClient struct {
	AgentAddr string
	ApiKey    string
}

func NewMCPClient(agentAddr, apiKey string) *MCPClient {
	return &MCPClient{
		AgentAddr: agentAddr,
		ApiKey:    apiKey,
	}
}

// GetSuggestion 获取ai 建议
func (c *MCPClient) GetSuggestion(ctx *types.AgentContext) (*types.FixSuggestion, error) {
	jsonData, err := json.Marshal(ctx)
	if err != nil {
		return nil, err
	}
	address := c.AgentAddr + "/api/v1/diagnose"
	request, err := http.NewRequest("POST", address, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("token", c.ApiKey)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败: %s", response.Status)
	}

	var suggestion types.FixSuggestion
	if err = json.NewDecoder(response.Body).Decode(&suggestion); err != nil {
		return nil, err
	}
	return &suggestion, nil
}
