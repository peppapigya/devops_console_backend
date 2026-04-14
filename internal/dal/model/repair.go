package model

import "time"

const TableNameRepairSession = "repair_sessions"

type RepairSession struct {
	ID               string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TraceID          string     `gorm:"column:trace_id;type:varchar(64);index" json:"trace_id"`
	Operator         string     `gorm:"column:operator;type:varchar(128)" json:"operator"`
	LogEventID       string     `gorm:"column:log_event_id;type:varchar(64);index" json:"log_event_id"`
	LogSource        string     `gorm:"column:log_source;type:varchar(255)" json:"log_source"`
	LogMessage       string     `gorm:"column:log_message;type:text" json:"log_message"`
	LogLevel         string     `gorm:"column:log_level;type:varchar(20)" json:"log_level"`
	LogHost          string     `gorm:"column:log_host;type:varchar(255)" json:"log_host"`
	LogService       string     `gorm:"column:log_service;type:varchar(255)" json:"log_service"`
	Analysis         string     `gorm:"column:analysis;type:text" json:"analysis"`
	RootCause        string     `gorm:"column:root_cause;type:text" json:"root_cause"`
	Severity         string     `gorm:"column:severity;type:varchar(20)" json:"severity"`
	Confidence       float64    `gorm:"column:confidence" json:"confidence"`
	Status           string     `gorm:"column:status;type:varchar(20);default:'created'" json:"status"`
	TotalActions     int        `gorm:"column:total_actions;default:0" json:"total_actions"`
	CompletedActions int        `gorm:"column:completed_actions;default:0" json:"completed_actions"`
	CreatedAt        *time.Time `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
	UpdatedAt        *time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updated_at"`
	FinishedAt       *time.Time `gorm:"column:finished_at;type:datetime(3)" json:"finished_at,omitempty"`
}

func (*RepairSession) TableName() string { return TableNameRepairSession }

const TableNameSessionMessage = "session_messages"

type SessionMessage struct {
	ID         uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	SessionID  string     `gorm:"column:session_id;type:varchar(36);not null;index" json:"session_id"`
	Role       string     `gorm:"column:role;type:varchar(20);not null" json:"role"`
	Content    string     `gorm:"column:content;type:mediumtext;not null" json:"content"`
	IsSummary  bool       `gorm:"column:is_summary;default:false" json:"is_summary"`
	TokenCount int        `gorm:"column:token_count;default:0" json:"token_count"`
	CreatedAt  *time.Time `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
}

func (*SessionMessage) TableName() string { return TableNameSessionMessage }

const TableNameRepairAction = "repair_actions"

type RepairAction struct {
	ID              uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	SessionID       string     `gorm:"column:session_id;type:varchar(36);not null;index" json:"session_id"`
	ActionOrder     int        `gorm:"column:action_order;not null" json:"action_order"`
	ToolName        string     `gorm:"column:tool_name;type:varchar(128)" json:"tool_name"`
	Description     string     `gorm:"column:description;type:varchar(512)" json:"description"`
	Thought         string     `gorm:"column:thought;type:text" json:"thought"`
	Command         string     `gorm:"column:command;type:text" json:"command"`
	Cwd             string     `gorm:"column:cwd;type:varchar(255);default:'/'" json:"cwd"`
	Target          string     `gorm:"column:target;type:varchar(255)" json:"target"`
	Timeout         int        `gorm:"column:timeout;default:30" json:"timeout"`
	RiskLevel       string     `gorm:"column:risk_level;type:varchar(20)" json:"risk_level"`
	RiskReason      string     `gorm:"column:risk_reason;type:text" json:"risk_reason"`
	RollbackCommand string     `gorm:"column:rollback_command;type:text" json:"rollback_command"`
	ApprovedBy      string     `gorm:"column:approved_by;type:varchar(128)" json:"approved_by"`
	ApprovedAt      *time.Time `gorm:"column:approved_at;type:datetime(3)" json:"approved_at,omitempty"`
	Status          string     `gorm:"column:status;type:varchar(20);default:'pending'" json:"status"`
	Output          string     `gorm:"column:output;type:mediumtext" json:"output"`
	ErrorMsg        string     `gorm:"column:error_msg;type:text" json:"error_msg"`
	ExitCode        int        `gorm:"column:exit_code;default:0" json:"exit_code"`
	DurationMs      int        `gorm:"column:duration_ms;default:0" json:"duration_ms"`
	ExecutedAt      *time.Time `gorm:"column:executed_at;type:datetime(3)" json:"executed_at,omitempty"`
	CreatedAt       *time.Time `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
}

func (*RepairAction) TableName() string { return TableNameRepairAction }

const TableNameRepairSessionEvent = "repair_session_events"

type RepairSessionEvent struct {
	ID        uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	SessionID string     `gorm:"column:session_id;type:varchar(36);not null;index" json:"session_id"`
	EventType string     `gorm:"column:event_type;type:varchar(64);not null" json:"event_type"`
	Payload   string     `gorm:"column:payload;type:mediumtext" json:"payload"`
	CreatedAt *time.Time `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
}

func (*RepairSessionEvent) TableName() string { return TableNameRepairSessionEvent }
