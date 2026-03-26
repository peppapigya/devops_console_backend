package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameTaskNode = "task_nodes"

// TaskNode 任务节点定义
type TaskNode struct {
	ID         uint32         `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	WorkflowID uint32         `gorm:"column:workflow_id;type:int unsigned;not null;index:idx_workflow_id;comment:关联工作流ID" json:"workflowId"`
	NodeName   string         `gorm:"column:node_name;type:varchar(128);not null;comment:节点名称" json:"nodeName"`
	NodeType   string         `gorm:"column:node_type;type:varchar(32);not null;comment:节点类型" json:"nodeType"`
	TaskType   string         `gorm:"column:task_type;type:varchar(32);not null;comment:任务类型" json:"taskType"`
	TargetID   uint64         `gorm:"column:target_id;comment:执行目标ID" json:"targetId"`
	TargetType string         `gorm:"column:target_type;type:varchar(20);comment:目标类型：host/k8s/db" json:"targetType"`
	Config     *string        `gorm:"column:config;type:json;comment:节点配置(JSON)" json:"config"`
	Timeout    int32          `gorm:"column:timeout;type:int;not null;default:0;comment:超时时间(秒)" json:"timeout"`
	RetryCount int32          `gorm:"column:retry_count;type:int;not null;default:0;comment:重试次数" json:"retryCount"`
	PositionX  float64        `gorm:"column:position_x;type:double;not null;default:0;comment:画布坐标X" json:"positionX"`
	PositionY  float64        `gorm:"column:position_y;type:double;not null;default:0;comment:画布坐标Y" json:"positionY"`
	SortOrder  int32          `gorm:"column:sort_order;type:int;not null;default:0;comment:排序" json:"sortOrder"`
	CreatedAt  *time.Time     `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;type:datetime" json:"deletedAt"`
}

// TableName TaskNode's table name
func (*TaskNode) TableName() string {
	return TableNameTaskNode
}
