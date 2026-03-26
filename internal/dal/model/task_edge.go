package model

import (
	"time"
)

const TableNameTaskEdge = "task_edges"

// TaskEdge 任务边/依赖关系
type TaskEdge struct {
	ID           uint32     `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	WorkflowID   uint32     `gorm:"column:workflow_id;type:int unsigned;not null;index:idx_workflow_id;comment:关联工作流ID" json:"workflowId"`
	SourceNodeID uint32     `gorm:"column:source_node_id;type:int unsigned;not null;comment:上游节点ID" json:"sourceNodeId"`
	TargetNodeID uint32     `gorm:"column:target_node_id;type:int unsigned;not null;comment:下游节点ID" json:"targetNodeId"`
	Condition    *string    `gorm:"column:condition;type:varchar(255);comment:执行条件" json:"condition"`
	EdgeType     string     `gorm:"column:edge_type;type:varchar(32);not null;default:default;comment:边类型" json:"edgeType"`
	SourceHandle string     `gorm:"column:source_handle;type:varchar(20)" json:"sourceHandle"`
	TargetHandle string     `gorm:"column:target_handle;type:varchar(20)" json:"targetHandle"`
	CreatedAt    *time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

// TableName TaskEdge's table name
func (*TaskEdge) TableName() string {
	return TableNameTaskEdge
}
