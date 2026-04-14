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

type LLMConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type LLMResponse struct {
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Content   string     `json:"content,omitempty"`
	IsFinish  bool
}

type LLMClient struct {
	cfg     LLMConfig
	tools   []openai.ChatCompletionToolUnionParam
	history []openai.ChatCompletionMessageParamUnion
}

func NewLLMClient(cfg LLMConfig) *LLMClient {
	return &LLMClient{
		cfg:     cfg,
		tools:   buildToolParams(),
		history: []openai.ChatCompletionMessageParamUnion{},
	}
}

func (c *LLMClient) SetSystemPrompt(prompt string) {
	c.history = append([]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(prompt)}, c.history...)
}

func (c *LLMClient) AddUserMessage(content string) {
	c.history = append(c.history, openai.UserMessage(content))
}

func (c *LLMClient) AddToolResult(toolCallID, content string) {
	c.history = append(c.history, openai.ToolMessage(toolCallID, content))
}

func (c *LLMClient) Call(ctx context.Context) (*LLMResponse, error) {
	client := openai.NewClient(option.WithAPIKey(c.cfg.APIKey), option.WithBaseURL(c.cfg.BaseURL))
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    c.cfg.Model,
		Messages: c.history,
		Tools:    c.tools,
	})
	if err != nil {
		logger.Errorf("call llm failed: %v", err)
		return nil, fmt.Errorf("llm call failed: %v", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("llm returned empty choices")
	}

	msg := resp.Choices[0].Message
	c.history = append(c.history, msg.ToParam())

	result := &LLMResponse{}
	if len(msg.ToolCalls) > 0 {
		for _, tc := range msg.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				args = map[string]interface{}{"raw": tc.Function.Arguments}
			}
			result.ToolCalls = append(result.ToolCalls, ToolCall{ID: tc.ID, Name: tc.Function.Name, Arguments: args})
		}
		return result, nil
	}

	result.Content = strings.TrimSpace(msg.Content)
	result.IsFinish = true
	return result, nil
}

func buildToolParams() []openai.ChatCompletionToolUnionParam {
	commonMeta := map[string]interface{}{
		"thought": map[string]interface{}{
			"type":        "string",
			"description": "中文推理说明：已掌握哪些证据、为何需要该工具、期望获得什么信息。",
		},
		"description": map[string]interface{}{
			"type":        "string",
			"description": "中文步骤摘要：面向用户展示本步骤要做什么。",
		},
		"risk_level": map[string]interface{}{
			"type":        "string",
			"description": "操作风险等级。",
			"enum":        []string{"low", "medium", "high"},
		},
		"risk_reason": map[string]interface{}{
			"type":        "string",
			"description": "中文说明：为什么该操作是这个风险等级。",
		},
		"host":     map[string]interface{}{"type": "string"},
		"port":     map[string]interface{}{"type": "integer", "default": 22},
		"user":     map[string]interface{}{"type": "string"},
		"password": map[string]interface{}{"type": "string"},
		"timeout":  map[string]interface{}{"type": "integer", "default": 30},
	}

	toolWithSSH := func(name, desc string, extra map[string]interface{}, required []string) openai.ChatCompletionToolUnionParam {
		properties := map[string]interface{}{}
		for k, v := range commonMeta {
			properties[k] = v
		}
		for k, v := range extra {
			properties[k] = v
		}
		baseRequired := []string{"thought", "description", "risk_level", "risk_reason", "host", "user", "password"}
		baseRequired = append(baseRequired, required...)
		return openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        name,
			Description: openai.String(desc),
			Parameters: openai.FunctionParameters{
				"type":       "object",
				"properties": properties,
				"required":   baseRequired,
			},
		})
	}

	reportFunc := openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        "submit_diagnosis_report",
		Description: openai.String("提交最终 RCA 报告。这是结束本次排障的唯一合法方式。报告字段请使用中文。"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"thought":        map[string]interface{}{"type": "string"},
				"summary":        map[string]interface{}{"type": "string"},
				"root_cause":     map[string]interface{}{"type": "string"},
				"severity":       map[string]interface{}{"type": "string", "enum": []string{"low", "medium", "high", "critical"}},
				"fixed":          map[string]interface{}{"type": "boolean"},
				"symptoms":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
				"evidence":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
				"fix_summary":    map[string]interface{}{"type": "string"},
				"recommendation": map[string]interface{}{"type": "string"},
				"next_steps":     map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			},
			"required": []string{"thought", "summary", "root_cause", "severity", "fixed", "symptoms", "evidence"},
		},
	})

	return []openai.ChatCompletionToolUnionParam{
		toolWithSSH("inspect_service_status", "检查 Linux 服务运行状态。", map[string]interface{}{
			"service": map[string]interface{}{"type": "string", "description": "Service name"},
		}, []string{"service"}),
		toolWithSSH("inspect_service_logs", "读取 Linux 服务最近日志。", map[string]interface{}{
			"service": map[string]interface{}{"type": "string", "description": "Service name"},
			"lines":   map[string]interface{}{"type": "integer", "default": 80},
		}, []string{"service"}),
		toolWithSSH("inspect_file_snippet", "读取文件指定行范围内容。", map[string]interface{}{
			"file_path":  map[string]interface{}{"type": "string"},
			"line_start": map[string]interface{}{"type": "integer"},
			"line_end":   map[string]interface{}{"type": "integer"},
		}, []string{"file_path", "line_start", "line_end"}),
		toolWithSSH("validate_nginx_config", "执行 nginx -t 并返回精确校验输出。", map[string]interface{}{
			"config_path": map[string]interface{}{"type": "string"},
		}, nil),
		toolWithSSH("replace_file_content", "替换文件中的精确文本片段。", map[string]interface{}{
			"file_path":     map[string]interface{}{"type": "string"},
			"search":        map[string]interface{}{"type": "string"},
			"replace":       map[string]interface{}{"type": "string"},
			"create_backup": map[string]interface{}{"type": "boolean", "default": true},
		}, []string{"file_path", "search", "replace"}),
		toolWithSSH("restart_service", "重启 Linux 服务。", map[string]interface{}{
			"service": map[string]interface{}{"type": "string", "description": "Service name"},
		}, []string{"service"}),
		toolWithSSH("execute_ssh", "结构化工具无法表达时的兜底 SSH 命令（优先只读）。", map[string]interface{}{
			"command": map[string]interface{}{"type": "string"},
			"cwd":     map[string]interface{}{"type": "string", "default": "/"},
		}, []string{"command"}),
		reportFunc,
	}
}
