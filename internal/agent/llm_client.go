package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

// LLMConfig LLM 连接配置
type LLMConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Message 对话历史单条消息（用于持久化）
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolCall LLM 返回的工具调用指令
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// LLMResponse LLM 本轮回复
type LLMResponse struct {
	// LLM 要求调用工具
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	// 文本回复（诊断结论/最终报告）
	Content  string `json:"content,omitempty"`
	IsFinish bool   // 是否已完成（没有工具调用，输出了最终内容）
}

// LLMClient 直接连接通义千问（OpenAI 兼容模式），支持 Function Calling
type LLMClient struct {
	cfg     LLMConfig
	tools   []openai.ChatCompletionToolUnionParam
	history []openai.ChatCompletionMessageParamUnion
}

// NewLLMClient 创建 LLM 客户端
func NewLLMClient(cfg LLMConfig) *LLMClient {
	return &LLMClient{
		cfg:     cfg,
		tools:   buildToolParams(),
		history: []openai.ChatCompletionMessageParamUnion{},
	}
}

// SetSystemPrompt 设置系统提示词（插到历史最前）
func (c *LLMClient) SetSystemPrompt(prompt string) {
	c.history = append(
		[]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(prompt)},
		c.history...,
	)
}

// AddUserMessage 追加用户消息
func (c *LLMClient) AddUserMessage(content string) {
	c.history = append(c.history, openai.UserMessage(content))
}

// AddToolResult 将工具执行结果追加到历史
func (c *LLMClient) AddToolResult(toolCallID, content string) {
	c.history = append(c.history, openai.ToolMessage(toolCallID, content))
}

// Call 调用 LLM，返回本轮回复（工具调用或最终答案）
func (c *LLMClient) Call(ctx context.Context) (*LLMResponse, error) {
	client := openai.NewClient(
		option.WithAPIKey(c.cfg.APIKey),
		option.WithBaseURL(c.cfg.BaseURL),
	)

	params := openai.ChatCompletionNewParams{
		Model:    c.cfg.Model,
		Messages: c.history,
		Tools:    c.tools,
	}

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		logger.Errorf("调用 LLM API 失败: %v", err)
		return nil, fmt.Errorf("LLM 调用失败: %v", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM 返回空 choices")
	}

	choice := resp.Choices[0]
	msg := choice.Message

	// 将 assistant 消息追加到历史
	c.history = append(c.history, msg.ToParam())

	result := &LLMResponse{}

	// 判断是否有工具调用
	if len(msg.ToolCalls) > 0 {
		for _, tc := range msg.ToolCalls {
			var args map[string]interface{}
			if parseErr := json.Unmarshal([]byte(tc.Function.Arguments), &args); parseErr != nil {
				args = map[string]interface{}{"raw": tc.Function.Arguments}
			}
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: args,
			})
		}
		return result, nil
	}

	// 文本回复
	result.Content = strings.TrimSpace(msg.Content)
	result.IsFinish = true
	return result, nil
}

// buildToolParams 构建 OpenAI Function Calling 格式的工具列表
func buildToolParams() []openai.ChatCompletionToolUnionParam {
	sshFunc := shared.FunctionDefinitionParam{
		Name:        "execute_ssh",
		Description: openai.String("在目标主机上执行一条 Shell 命令，返回 stdout、stderr 和退出码。用于诊断系统状态或执行修复操作。每次只执行一条命令，等待结果后再决策。"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"host": map[string]interface{}{
					"type":        "string",
					"description": "目标主机 IP 地址或域名",
				},
				"port": map[string]interface{}{
					"type":        "integer",
					"description": "SSH 端口号，默认 22",
					"default":     22,
				},
				"user": map[string]interface{}{
					"type":        "string",
					"description": "SSH 登录用户名",
				},
				"password": map[string]interface{}{
					"type":        "string",
					"description": "SSH 登录密码",
				},
				"command": map[string]interface{}{
					"type":        "string",
					"description": "要执行的单条 Shell 命令",
				},
				"cwd": map[string]interface{}{
					"type":        "string",
					"description": "命令执行的工作目录，默认为 /",
					"default":     "/",
				},
				"timeout": map[string]interface{}{
					"type":        "integer",
					"description": "命令执行超时时间（秒），默认 30",
					"default":     30,
				},
				"risk_level": map[string]interface{}{
					"type":        "string",
					"description": "命令风险等级：low（纯读写查询）、medium（状态变更）、high（重启/删除/修改关键配置）",
					"enum":        []string{"low", "medium", "high"},
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "该命令的用途说明（将显示给用户）",
				},
			},
			"required": []string{"host", "user", "password", "command", "risk_level", "description"},
		},
	}
	return []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(sshFunc),
	}
}
