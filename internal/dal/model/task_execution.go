package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameTaskExecution = "task_executions"

// TaskExecution 任务执行记录
type TaskExecution struct {
	ID            uint32         `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	WorkflowID    uint32         `gorm:"column:workflow_id;type:int unsigned;not null;index:idx_workflow_id;comment:关联工作流ID" json:"workflowId"`
	ExecutionNo   string         `gorm:"column:execution_no;type:varchar(64);not null;index:idx_execution_no;comment:执行编号" json:"executionNo"`
	TriggerType   string         `gorm:"column:trigger_type;type:varchar(32);not null;comment:触发类型" json:"triggerType"`
	Status        string         `gorm:"column:status;type:varchar(32);not null;comment:执行状态" json:"status"`
	StartTime     *time.Time     `gorm:"column:start_time;type:timestamp;comment:开始时间" json:"startTime"`
	EndTime       *time.Time     `gorm:"column:end_time;type:timestamp;comment:结束时间" json:"endTime"`
	Duration      int32          `gorm:"column:duration;type:int;comment:耗时(秒)" json:"duration"`
	Result        *string        `gorm:"column:result;type:json;comment:执行结果(JSON)" json:"result"`
	ErrorMessage  *string        `gorm:"column:error_message;type:text;comment:错误信息" json:"errorMessage"`
	TriggeredBy   uint32         `gorm:"column:triggered_by;type:int unsigned;comment:触发人ID" json:"triggeredBy"`
	RetryCount    int32          `gorm:"column:retry_count;type:int;not null;default:0;comment:重试次数" json:"retryCount"`
	NextRetryTime *time.Time     `gorm:"column:next_retry_time;type:timestamp;comment:下一次重试时间" json:"nextRetryTime"`
	CreatedAt     *time.Time     `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt     *time.Time     `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;type:datetime" json:"deletedAt"`
}

// TableName TaskExecution's table name
func (*TaskExecution) TableName() string {
	return TableNameTaskExecution
}
