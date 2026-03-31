package types

import "time"

// AgentContext 表示 Agent 处理上下文
// 包含进行分析和修复所需的所有信息
type AgentContext struct {
	LogEvent      LogEvent    `json:"log_event"`      // 触发的日志事件
	Rule          MonitorRule `json:"rule"`           // 匹配的规则
	RelatedLogs   []LogEvent  `json:"related_logs"`   // 相关日志（上下文）
	HostTarget    *string     `json:"host_target"`    // 远程目标（如果是远程修复）
	AttemptCount  int         `json:"attempt_count"`  // 当前是第几次尝试修复
	PreviousFixes []FixResult `json:"previous_fixes"` // 之前的修复尝试结果
	Timestamp     time.Time   `json:"timestamp"`      // 上下文创建时间
}

// LogEvent 表示一个日志事件
// 包含日志的完整上下文信息（来源、级别、内容等）
type LogEvent struct {
	ID        string                 `json:"id"`        // 事件唯一标识
	Timestamp time.Time              `json:"timestamp"` // 事件时间戳
	Source    string                 `json:"source"`    // 日志来源（如服务名、文件路径）
	Level     string                 `json:"level"`     // 日志级别（ERROR, WARN, INFO等）
	Message   string                 `json:"message"`   // 日志消息内容
	Host      string                 `json:"host"`      // 产生日志的主机
	Service   string                 `json:"service"`   // 产生日志的服务名称
	Metadata  map[string]interface{} `json:"metadata"`  // 额外的元数据信息
}

type FixResult struct {
	PlanID     string         `json:"plan_id"`     // 关联的计划ID
	Success    bool           `json:"success"`     // 是否成功
	Actions    []ActionResult `json:"actions"`     // 各个动作的执行结果
	Duration   int            `json:"duration"`    // 执行耗时（秒）
	NewLogs    []LogEvent     `json:"new_logs"`    // 执行后产生的新日志
	ErrorMsg   string         `json:"error_msg"`   // 错误信息（如失败）
	ExecutedAt time.Time      `json:"executed_at"` // 执行时间
	ExecutedBy string         `json:"executed_by"` // 执行者（local/remote）
	RemoteHost string         `json:"remote_host"` // 远程主机（如适用）
}

// ActionResult 表示单个动作的执行结果
type ActionResult struct {
	Action     string    `json:"action"`      // 动作描述
	Success    bool      `json:"success"`     // 是否成功
	Output     string    `json:"output"`      // 执行输出
	Error      string    `json:"error"`       // 错误信息
	ExitCode   int       `json:"exit_code"`   // 退出码
	Duration   int       `json:"duration"`    // 耗时（毫秒）
	ExecutedAt time.Time `json:"executed_at"` // 执行时间
}

// MonitorRule 定义监控规则
// 用于判断哪些日志事件需要触发自动修复流程
type MonitorRule struct {
	ID          string   `json:"id"`          // 规则唯一标识
	Name        string   `json:"name"`        // 规则名称
	Enabled     bool     `json:"enabled"`     // 是否启用
	Level       string   `json:"level"`       // 触发的日志级别（ERROR, WARN等）
	Keywords    []string `json:"keywords"`    // 关键词列表（OR关系）
	Pattern     string   `json:"pattern"`     // 正则表达式模式
	Service     string   `json:"service"`     // 目标服务名（空表示所有）
	Threshold   int      `json:"threshold"`   // 触发阈值（在时间窗口内出现次数）
	TimeWindow  int      `json:"time_window"` // 时间窗口（秒）
	Description string   `json:"description"` // 规则描述
}

// FixSuggestion 表示 LLM 返回的修复建议（必须是结构化JSON）
type FixSuggestion struct {
	Analysis      string      `json:"analysis"`       // 问题分析
	RootCause     string      `json:"root_cause"`     // 根因判断
	Severity      string      `json:"severity"`       // 严重程度（low, medium, high, critical）
	FixActions    []FixAction `json:"fix_actions"`    // 修复动作列表
	RiskLevel     string      `json:"risk_level"`     // 风险等级（low, medium, high）
	Rollbackable  bool        `json:"rollbackable"`   // 是否可回滚
	EstimatedTime int         `json:"estimated_time"` // 预计修复时间（秒）
	Confidence    float64     `json:"confidence"`     // AI置信度（0-1）
}

// FixAction 表示单个修复动作
type FixAction struct {
	Thought         string            `json:"thought"`          // （必填）分析前因后果和推理逻辑
	Type            string            `json:"type"`             // 动作类型（execute_cmd等）
	Target          string            `json:"target"`           // 目标（主机IP、服务名等）
	Parameters      map[string]string `json:"parameters"`       // 参数
	Command         string            `json:"command"`          // 实际执行的单条 shell 命令
	Cwd             string            `json:"cwd"`              // 执行的工作目录，默认为 /
	Timeout         int               `json:"timeout"`          // 预计超时时间秒数，默认 30
	RiskLevel       string            `json:"risk_level"`       // 风险评估（low, medium, high）
	RiskReason      string            `json:"risk_reason"`      // 具体的潜在影响分析
	RollbackCommand string            `json:"rollback_command"` // 修改操作的恢复命令，如果是查询则留空
	Description     string            `json:"description"`      // 动作描述
	Order           int               `json:"order"`            // 执行顺序
}
