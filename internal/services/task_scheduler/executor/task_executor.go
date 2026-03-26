package executor

import "context"

// TaskExecutionContext 执行上下文
type TaskExecutionContext struct {
	ExecutionID uint64
	NodeID      uint64
	WorkflowID  uint64
	Config      map[string]interface{}
	TargetID    uint64
	TargetType  string
	Logger      ExecutionLogger
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success    bool
	Output     string
	ErrorMsg   string
	Duration   int64 // 毫秒
	OutputVars map[string]interface{}
}

// ExecutionLogger 执行日志接口
type ExecutionLogger interface {
	Log(level string, message string)
	GetLogs() string
}

// TaskExecutor 任务执行器接口（策略模式）
type TaskExecutor interface {
	// Execute 执行任务
	Execute(ctx context.Context, context *TaskExecutionContext) *ExecutionResult
	// Validate 验证配置
	Validate(config map[string]interface{}) error
	// GetType 获取执行器类型
	GetType() string
}
