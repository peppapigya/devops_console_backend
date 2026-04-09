package model

import "time"

// ===================== 根因分析会话 =====================

const TableNameRepairSession = "repair_sessions"

// RepairSession 根因分析与修复会话主表
type RepairSession struct {
	ID string `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	// 触发告警的日志事件信息（冗余存储，避免 JOIN）
	LogEventID string `gorm:"column:log_event_id;type:varchar(64);index" json:"log_event_id"`
	LogSource  string `gorm:"column:log_source;type:varchar(255)" json:"log_source"`
	LogMessage string `gorm:"column:log_message;type:text" json:"log_message"`
	LogLevel   string `gorm:"column:log_level;type:varchar(20)" json:"log_level"`
	LogHost    string `gorm:"column:log_host;type:varchar(255)" json:"log_host"`
	LogService string `gorm:"column:log_service;type:varchar(255)" json:"log_service"`
	// AI 分析输出
	Analysis   string  `gorm:"column:analysis;type:text" json:"analysis"`
	RootCause  string  `gorm:"column:root_cause;type:text" json:"root_cause"`
	Severity   string  `gorm:"column:severity;type:varchar(20)" json:"severity"` // low|medium|high|critical
	Confidence float64 `gorm:"column:confidence" json:"confidence"`
	// 执行状态
	Status           string `gorm:"column:status;type:varchar(20);default:'pending'" json:"status"` // pending|running|waiting_confirm|success|failed|paused
	TotalActions     int    `gorm:"column:total_actions;default:0" json:"total_actions"`
	CompletedActions int    `gorm:"column:completed_actions;default:0" json:"completed_actions"`
	// 时间
	CreatedAt  *time.Time `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updated_at"`
	FinishedAt *time.Time `gorm:"column:finished_at;type:datetime(3)" json:"finished_at,omitempty"`
}

func (*RepairSession) TableName() string { return TableNameRepairSession }

// ===================== 对话消息记录 =====================

const TableNameSessionMessage = "session_messages"

// SessionMessage 每轮 AI 对话消息，支持 Token 压缩后的摘要替换
type SessionMessage struct {
	ID         uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	SessionID  string     `gorm:"column:session_id;type:varchar(36);not null;index" json:"session_id"`
	Role       string     `gorm:"column:role;type:varchar(20);not null" json:"role"` // system|user|assistant
	Content    string     `gorm:"column:content;type:mediumtext;not null" json:"content"`
	IsSummary  bool       `gorm:"column:is_summary;default:false" json:"is_summary"` // 是否为压缩摘要行
	TokenCount int        `gorm:"column:token_count;default:0" json:"token_count"`
	CreatedAt  *time.Time `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
}

func (*SessionMessage) TableName() string { return TableNameSessionMessage }

// ===================== 修复动作执行记录 =====================

const TableNameRepairAction = "repair_actions"

// RepairAction 修复计划中每个动作的执行详情
type RepairAction struct {
	ID              uint64 `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	SessionID       string `gorm:"column:session_id;type:varchar(36);not null;index" json:"session_id"`
	ActionOrder     int    `gorm:"column:action_order;not null" json:"action_order"`
	Description     string `gorm:"column:description;type:varchar(512)" json:"description"`
	Thought         string `gorm:"column:thought;type:text" json:"thought"`
	Command         string `gorm:"column:command;type:text" json:"command"`
	Cwd             string `gorm:"column:cwd;type:varchar(255);default:'/'" json:"cwd"`
	Target          string `gorm:"column:target;type:varchar(255)" json:"target"` // SSH 目标主机IP
	Timeout         int    `gorm:"column:timeout;default:30" json:"timeout"`
	RiskLevel       string `gorm:"column:risk_level;type:varchar(20)" json:"risk_level"` // low|medium|high
	RiskReason      string `gorm:"column:risk_reason;type:text" json:"risk_reason"`
	RollbackCommand string `gorm:"column:rollback_command;type:text" json:"rollback_command"`
	// 执行结果
	Status     string     `gorm:"column:status;type:varchar(20);default:'pending'" json:"status"` // pending|waiting_confirm|running|success|failed|skipped
	Output     string     `gorm:"column:output;type:mediumtext" json:"output"`
	ErrorMsg   string     `gorm:"column:error_msg;type:text" json:"error_msg"`
	ExitCode   int        `gorm:"column:exit_code;default:0" json:"exit_code"`
	DurationMs int        `gorm:"column:duration_ms;default:0" json:"duration_ms"`
	ExecutedAt *time.Time `gorm:"column:executed_at;type:datetime(3)" json:"executed_at,omitempty"`
	CreatedAt  *time.Time `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
}

func (*RepairAction) TableName() string { return TableNameRepairAction }
