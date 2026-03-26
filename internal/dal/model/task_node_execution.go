package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameTaskNodeExecution = "task_node_executions"

// TaskNodeExecution 任务节点执行记录
type TaskNodeExecution struct {
	ID           uint32         `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	ExecutionID  uint32         `gorm:"column:execution_id;type:int unsigned;not null;index:idx_execution_id;comment:关联任务执行ID" json:"executionId"`
	NodeID       uint32         `gorm:"column:node_id;type:int unsigned;not null;comment:关联节点ID" json:"nodeId"`
	Status       string         `gorm:"column:status;type:varchar(32);not null;comment:节点执行状态" json:"status"`
	StartTime    *time.Time     `gorm:"column:start_time;type:timestamp;comment:开始时间" json:"startTime"`
	EndTime      *time.Time     `gorm:"column:end_time;type:timestamp;comment:结束时间" json:"endTime"`
	Duration     int32          `gorm:"column:duration;type:int;comment:耗时(秒)" json:"duration"`
	InputParams  *string        `gorm:"column:input_params;type:json;comment:输入参数(JSON)" json:"inputParams"`
	OutputParams *string        `gorm:"column:output_params;type:json;comment:输出参数(JSON)" json:"outputParams"`
	Logs         *string        `gorm:"column:logs;type:longtext;comment:执行日志" json:"logs"`
	ErrorMessage *string        `gorm:"column:error_message;type:text;comment:错误信息" json:"errorMessage"`
	RetryCount   int32          `gorm:"column:retry_count;type:int;not null;default:0;comment:重试次数" json:"retryCount"`
	CreatedAt    *time.Time     `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt    *time.Time     `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;type:datetime" json:"deletedAt"`
}

// TableName TaskNodeExecution's table name
func (*TaskNodeExecution) TableName() string {
	return TableNameTaskNodeExecution
}
