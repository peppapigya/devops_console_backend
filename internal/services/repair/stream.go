package repair

import (
	"encoding/json"
	"fmt"
	"sync"

	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
)

// SSEEventType SSE 事件类型
type SSEEventType string

const (
	EventThinking      SSEEventType = "thinking"       // AI 正在思考（打字机流式输出）
	EventPlan          SSEEventType = "plan"           // AI 输出了修复计划
	EventActionStart   SSEEventType = "action_start"   // 开始执行某个动作
	EventActionOutput  SSEEventType = "action_output"  // 动作实时输出（流式日志）
	EventActionResult  SSEEventType = "action_result"  // 动作执行完成
	EventWaitConfirm   SSEEventType = "wait_confirm"   // 高风险动作等待用户确认
	EventSessionUpdate SSEEventType = "session_update" // session 状态更新
	EventDone          SSEEventType = "done"           // 全部完成
	EventError         SSEEventType = "error"          // 出错
)

// SSEEvent SSE 推送事件结构
type SSEEvent struct {
	Type      SSEEventType `json:"type"`
	SessionID string       `json:"session_id"`
	Payload   interface{}  `json:"payload,omitempty"`
}

// ThinkingPayload 思考过程 payload
type ThinkingPayload struct {
	Content string `json:"content"` // 流式 token
}

// PlanPayload 修复计划 payload
type PlanPayload struct {
	Analysis   string        `json:"analysis"`
	RootCause  string        `json:"root_cause"`
	Severity   string        `json:"severity"`
	Confidence float64       `json:"confidence"`
	Actions    []ActionBrief `json:"actions"`
}

// ActionBrief 动作摘要（计划阶段展示，不含执行结果）
type ActionBrief struct {
	ID          uint64 `json:"id"`
	Order       int    `json:"order"`
	Description string `json:"description"`
	Command     string `json:"command"`
	RiskLevel   string `json:"risk_level"`
	RiskReason  string `json:"risk_reason"`
}

// ActionStartPayload 动作开始 payload
type ActionStartPayload struct {
	ActionID    uint64 `json:"action_id"`
	ActionOrder int    `json:"action_order"`
	ToolName    string `json:"tool_name"`
	Thought     string `json:"thought"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Target      string `json:"target"`
}

// ActionOutputPayload 动作实时输出 payload
type ActionOutputPayload struct {
	ActionID uint64 `json:"action_id"`
	Line     string `json:"line"` // 一行输出
}

// ActionResultPayload 动作完成 payload
type ActionResultPayload struct {
	ActionID   uint64 `json:"action_id"`
	Status     string `json:"status"` // success|failed|skipped
	Output     string `json:"output"`
	ErrorMsg   string `json:"error_msg"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int    `json:"duration_ms"`
}

// WaitConfirmPayload 等待确认 payload
type WaitConfirmPayload struct {
	ActionID    uint64 `json:"action_id"`
	ActionOrder int    `json:"action_order"`
	Description string `json:"description"`
	Command     string `json:"command"`
	RiskReason  string `json:"risk_reason"`
}

// SessionUpdatePayload session 状态更新 payload
type SessionUpdatePayload struct {
	Status           string `json:"status"`
	CompletedActions int    `json:"completed_actions"`
	TotalActions     int    `json:"total_actions"`
}

// DonePayload 完成 payload
type DonePayload struct {
	Status           string `json:"status"` // success|failed
	CompletedActions int    `json:"completed_actions"`
	TotalActions     int    `json:"total_actions"`
}

// ErrorPayload 错误 payload
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ============================================================
// StreamHub SSE 频道管理器：管理所有活跃 session 的订阅连接
// ============================================================

type StreamHub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan string // sessionID -> []channel
	eventMapper *mapper.RepairSessionEventMapper
}

var globalHub = &StreamHub{
	subscribers: make(map[string][]chan string),
}

// GetHub 获取全局 StreamHub 单例
func GetHub() *StreamHub {
	return globalHub
}

func (h *StreamHub) SetEventMapper(eventMapper *mapper.RepairSessionEventMapper) {
	h.eventMapper = eventMapper
}

// Subscribe 订阅一个 session 的 SSE 流，返回消息 channel 和取消函数
func (h *StreamHub) Subscribe(sessionID string) (chan string, func()) {
	ch := make(chan string, 100) // 带缓冲，防止慢消费导致 goroutine 泄露
	h.mu.Lock()
	h.subscribers[sessionID] = append(h.subscribers[sessionID], ch)
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		chans := h.subscribers[sessionID]
		newChans := make([]chan string, 0, len(chans)-1)
		for _, c := range chans {
			if c != ch {
				newChans = append(newChans, c)
			}
		}
		if len(newChans) == 0 {
			delete(h.subscribers, sessionID)
		} else {
			h.subscribers[sessionID] = newChans
		}
		close(ch)
	}
	return ch, cancel
}

// Publish 向 session 的所有订阅者广播一条 SSE 消息
func (h *StreamHub) Publish(event SSEEvent) {
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return
	}

	line := ""
	if h.eventMapper != nil {
		nowEvent := &model.RepairSessionEvent{
			SessionID: event.SessionID,
			EventType: string(event.Type),
			Payload:   string(payloadBytes),
		}
		if err := h.eventMapper.Create(nowEvent); err != nil {
			return
		}
		line = formatPersistedEvent(nowEvent)
	} else {
		data, err := json.Marshal(event)
		if err != nil {
			return
		}
		line = fmt.Sprintf("data: %s\n\n", string(data))
	}

	h.mu.RLock()
	chans := h.subscribers[event.SessionID]
	h.mu.RUnlock()

	for _, ch := range chans {
		select {
		case ch <- line:
		default:
			// 如果 channel 满了直接丢弃（客户端消费太慢），不阻塞业务
		}
	}
}

// PublishDone 发布完成事件并关闭所有订阅
func (h *StreamHub) PublishDone(sessionID string, payload DonePayload) {
	h.Publish(SSEEvent{Type: EventDone, SessionID: sessionID, Payload: payload})
	// 通知所有 channel 关闭
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.subscribers[sessionID] {
		// 发送关闭信号行（前端检测到后断开）
		select {
		case ch <- "event: close\ndata: {}\n\n":
		default:
		}
	}
}

func (h *StreamHub) Replay(sessionID string, sinceID uint64) ([]string, uint64, error) {
	if h.eventMapper == nil {
		return nil, sinceID, nil
	}
	events, err := h.eventMapper.ListBySessionIDSince(sessionID, sinceID, 1000)
	if err != nil {
		return nil, sinceID, err
	}
	lines := make([]string, 0, len(events))
	lastID := sinceID
	for i := range events {
		lines = append(lines, formatPersistedEvent(&events[i]))
		lastID = events[i].ID
	}
	return lines, lastID, nil
}

func formatPersistedEvent(event *model.RepairSessionEvent) string {
	payload := event.Payload
	if payload == "" {
		payload = "{}"
	}
	return fmt.Sprintf("id: %d\ndata: {\"type\":\"%s\",\"session_id\":\"%s\",\"payload\":%s}\n\n", event.ID, event.EventType, event.SessionID, payload)
}
