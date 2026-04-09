package repair

import (
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/types"
	"devops-console-backend/pkg/utils/logs"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SessionService 管理 repair session 的完整生命周期
type SessionService struct {
	sessionMapper *mapper.RepairSessionMapper
	msgMapper     *mapper.SessionMessageMapper
	actionMapper  *mapper.RepairActionMapper
	hub           *StreamHub
	mcpAgentURL   string
}

func NewSessionService(
	sessionMapper *mapper.RepairSessionMapper,
	msgMapper *mapper.SessionMessageMapper,
	actionMapper *mapper.RepairActionMapper,
	mcpAgentURL string,
) *SessionService {
	return &SessionService{
		sessionMapper: sessionMapper,
		msgMapper:     msgMapper,
		actionMapper:  actionMapper,
		hub:           GetHub(),
		mcpAgentURL:   mcpAgentURL,
	}
}

// CreateRequest 创建 session 的请求参数
type CreateRequest struct {
	LogEvent types.LogEvent `json:"log_event"`
	SSHUser  string         `json:"ssh_user"`
	SSHPass  string         `json:"ssh_pass"`
	SSHPort  int            `json:"ssh_port"`
}

// CreateSession 创建一个新的根因分析 session，返回 session_id
func (s *SessionService) CreateSession(req CreateRequest) (string, error) {
	id := uuid.New().String()
	now := time.Now()
	session := &model.RepairSession{
		ID:         id,
		LogEventID: req.LogEvent.ID,
		LogSource:  req.LogEvent.Source,
		LogMessage: req.LogEvent.Message,
		LogLevel:   req.LogEvent.Level,
		LogHost:    req.LogEvent.Host,
		LogService: req.LogEvent.Service,
		Status:     "pending",
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}
	if err := s.sessionMapper.Create(session); err != nil {
		return "", fmt.Errorf("创建 session 失败: %v", err)
	}
	return id, nil
}

// ──────────────────────────────────────────────────────────────
// AgentRunFunc 是 agent 包的 RunAgentLoop 函数签名
// 通过依赖注入避免循环引用（repair ← agent ← repair）
// ──────────────────────────────────────────────────────────────

// AgentSSHCreds SSH 凭据
type AgentSSHCreds struct {
	User     string
	Password string
	Port     int
}

// AgentLoopCfg Agent 运行配置（由调用方从 agent 包填充后传入）
type AgentLoopCfg struct {
	LLMAPIKey    string
	LLMBaseURL   string
	LLMModel     string
	MCPServerURL string
	MCPToken     string
	MaxRounds    int
}

// AgentRunFunc 函数类型，由 Controller 层将 agent.RunAgentLoop 注入进来
type AgentRunFunc func(
	sessionID string,
	logMessage, logHost, logService, logLevel string,
	creds AgentSSHCreds,
	cfg AgentLoopCfg,
	sessionMapper *mapper.RepairSessionMapper,
	msgMapper *mapper.SessionMessageMapper,
	actionMapper *mapper.RepairActionMapper,
	hub *StreamHub,
) error

// StartAsync 异步启动根因分析与修复流程
func (s *SessionService) StartAsync(sessionID string, req CreateRequest, cfg AgentLoopCfg, runFn AgentRunFunc) {
	go func() {
		err := runFn(
			sessionID,
			req.LogEvent.Message,
			req.LogEvent.Host,
			req.LogEvent.Service,
			req.LogEvent.Level,
			AgentSSHCreds{
				User:     req.SSHUser,
				Password: req.SSHPass,
				Port:     req.SSHPort,
			},
			cfg,
			s.sessionMapper,
			s.msgMapper,
			s.actionMapper,
			s.hub,
		)
		if err != nil {
			logs.Error(map[string]interface{}{}, fmt.Sprintf("session %s 执行失败: %v", sessionID, err))
			s.hub.Publish(SSEEvent{
				Type:      EventError,
				SessionID: sessionID,
				Payload:   ErrorPayload{Code: "RUN_ERROR", Message: err.Error()},
			})
			_ = s.sessionMapper.UpdateStatus(sessionID, "failed")
			now := time.Now()
			s.hub.PublishDone(sessionID, DonePayload{Status: "failed", CompletedActions: 0})
			_ = s.sessionMapper.UpdateFields(sessionID, map[string]interface{}{"finished_at": &now})
		}
	}()
}

// ============================================================
// confirmChannels 用于存储等待确认的 channel（全局单例）
// ============================================================

var confirmChannels = newConfirmStore()

type confirmStore struct {
	channels map[string]chan bool
	mu       *lockMU
}

type lockMU struct {
	ch chan struct{}
}

func newLock() *lockMU    { return &lockMU{ch: make(chan struct{}, 1)} }
func (l *lockMU) Lock()   { l.ch <- struct{}{} }
func (l *lockMU) Unlock() { <-l.ch }

func newConfirmStore() *confirmStore {
	return &confirmStore{
		channels: make(map[string]chan bool),
		mu:       newLock(),
	}
}

func (cs *confirmStore) key(sessionID string, actionID uint64) string {
	return fmt.Sprintf("%s:%d", sessionID, actionID)
}

// RegisterConfirmWait 注册等待确认的 channel（由 agent 包调用）
func RegisterConfirmWait(sessionID string, actionID uint64) chan bool {
	k := confirmChannels.key(sessionID, actionID)
	ch := make(chan bool, 1)
	confirmChannels.mu.Lock()
	confirmChannels.channels[k] = ch
	confirmChannels.mu.Unlock()
	return ch
}

// ConfirmAction 外部（Controller）调用，向等待中的动作发送确认信号
func ConfirmAction(sessionID string, actionID uint64, approved bool) bool {
	k := confirmChannels.key(sessionID, actionID)
	confirmChannels.mu.Lock()
	ch, ok := confirmChannels.channels[k]
	confirmChannels.mu.Unlock()
	if !ok {
		return false
	}
	ch <- approved
	confirmChannels.mu.Lock()
	delete(confirmChannels.channels, k)
	confirmChannels.mu.Unlock()
	return true
}
