package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameTaskWorkflow = "task_workflows"

// TaskWorkflow 任务工作流定义
type TaskWorkflow struct {
	ID              uint32         `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	Name            string         `gorm:"column:name;type:varchar(128);not null;comment:工作流名称" json:"name"`
	Description     *string        `gorm:"column:description;type:varchar(255);comment:工作流描述" json:"description"`
	CronExpression  *string        `gorm:"column:cron_expression;type:varchar(64);comment:Cron表达式" json:"cronExpression"`
	Status          int32          `gorm:"column:status;type:tinyint;not null;default:1;comment:状态: 1-启用, 2-禁用" json:"status"`
	TaskType        string         `gorm:"column:task_type;type:varchar(32);not null;comment:任务类型" json:"taskType"`
	RetryCount      int32          `gorm:"column:retry_count;type:int;not null;default:0;comment:重试次数" json:"retryCount"`
	RetryInterval   int32          `gorm:"column:retry_interval;type:int;not null;default:0;comment:重试间隔(秒)" json:"retryInterval"`
	Timeout         int32          `gorm:"column:timeout;type:int;not null;default:0;comment:超时时间(秒)" json:"timeout"`
	ConcurrentLimit int32          `gorm:"column:concurrent_limit;type:int;not null;default:1;comment:并发限制" json:"concurrentLimit"`
	CreatedBy       uint32         `gorm:"column:created_by;type:int unsigned;not null;comment:创建人ID" json:"createdBy"`
	CreatedAt       *time.Time     `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt       *time.Time     `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;type:datetime" json:"deletedAt"`
}

// TableName TaskWorkflow's table name
func (*TaskWorkflow) TableName() string {
	return TableNameTaskWorkflow
}
