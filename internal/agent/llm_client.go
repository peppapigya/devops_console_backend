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
			"description": "Reason about what evidence you already have, why this tool is needed, and what you expect to learn.",
		},
		"description": map[string]interface{}{
			"type":        "string",
			"description": "Short user-facing summary of this step.",
		},
		"risk_level": map[string]interface{}{
			"type":        "string",
			"description": "Risk level for the action.",
			"enum":        []string{"low", "medium", "high"},
		},
		"risk_reason": map[string]interface{}{
			"type":        "string",
			"description": "Why this action has the chosen risk level.",
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
		Description: openai.String("Submit the final RCA report. This is the only valid way to end the investigation."),
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
		toolWithSSH("inspect_service_status", "Inspect the runtime status of a Linux service.", map[string]interface{}{
			"service": map[string]interface{}{"type": "string", "description": "Service name"},
		}, []string{"service"}),
		toolWithSSH("inspect_service_logs", "Fetch recent logs for a Linux service.", map[string]interface{}{
			"service": map[string]interface{}{"type": "string", "description": "Service name"},
			"lines":   map[string]interface{}{"type": "integer", "default": 80},
		}, []string{"service"}),
		toolWithSSH("inspect_file_snippet", "Read a focused line range from a file.", map[string]interface{}{
			"file_path":  map[string]interface{}{"type": "string"},
			"line_start": map[string]interface{}{"type": "integer"},
			"line_end":   map[string]interface{}{"type": "integer"},
		}, []string{"file_path", "line_start", "line_end"}),
		toolWithSSH("validate_nginx_config", "Run nginx -t and capture the exact validator output.", map[string]interface{}{
			"config_path": map[string]interface{}{"type": "string"},
		}, nil),
		toolWithSSH("replace_file_content", "Replace one exact text fragment inside a file.", map[string]interface{}{
			"file_path":     map[string]interface{}{"type": "string"},
			"search":        map[string]interface{}{"type": "string"},
			"replace":       map[string]interface{}{"type": "string"},
			"create_backup": map[string]interface{}{"type": "boolean", "default": true},
		}, []string{"file_path", "search", "replace"}),
		toolWithSSH("restart_service", "Restart a Linux service.", map[string]interface{}{
			"service": map[string]interface{}{"type": "string", "description": "Service name"},
		}, []string{"service"}),
		toolWithSSH("execute_ssh", "Fallback raw SSH command when no typed read-only tool can express the inspection step.", map[string]interface{}{
			"command": map[string]interface{}{"type": "string"},
			"cwd":     map[string]interface{}{"type": "string", "default": "/"},
		}, []string{"command"}),
		reportFunc,
	}
}
