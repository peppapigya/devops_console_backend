package response

import (
	"encoding/json"
	"time"
)

// TaskWorkflowVO 工作流视图对象
type TaskWorkflowVO struct {
	ID             uint64       `json:"id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	CronExpression string       `json:"cronExpression"`
	Status         int          `json:"status"`
	CreatedAt      *time.Time   `json:"createdAt"`
	UpdatedAt      *time.Time   `json:"updatedAt"`
	Nodes          []TaskNodeVO `json:"nodes"`
	Edges          []TaskEdgeVO `json:"edges"`
}

// TaskWorkflowListResponse 工作流列表响应
type TaskWorkflowListResponse struct {
	Total int64            `json:"total"`
	List  []TaskWorkflowVO `json:"list"`
}

// TaskNodeVO 节点视图对象
type TaskNodeVO struct {
	ID         uint64          `json:"id"`
	WorkflowID uint64          `json:"workflowId"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Config     json.RawMessage `json:"config"`
	PositionX  float64         `json:"position_x"`
	PositionY  float64         `json:"position_y"`
	CreatedAt  *time.Time      `json:"createdAt"`
}

// TaskEdgeVO 边视图对象
type TaskEdgeVO struct {
	ID           uint64 `json:"id"`
	WorkflowID   uint64 `json:"workflowId"`
	FromNodeID   uint64 `json:"fromNodeId"`
	ToNodeID     uint64 `json:"toNodeId"`
	SourceHandle string `json:"sourceHandle"`
	TargetHandle string `json:"targetHandle"`
	Condition    string `json:"condition"`
}

// TaskExecutionVO 执行记录视图对象
type TaskExecutionVO struct {
	ID          uint64     `json:"id"`
	WorkflowID  uint64     `json:"workflowId"`
	Status      string     `json:"status"`
	StartTime   *time.Time `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
	Duration    int64      `json:"duration"` // Duration in milliseconds
	TriggerType string     `json:"triggerType"`
	TriggeredBy string     `json:"triggeredBy"`
}

// TaskExecutionListResponse 执行记录列表
type TaskExecutionListResponse struct {
	Total int64             `json:"total"`
	List  []TaskExecutionVO `json:"list"`
}

// TaskNodeExecutionVO 节点执行记录视图对象
type TaskNodeExecutionVO struct {
	ID          uint64     `json:"id"`
	ExecutionID uint64     `json:"executionId"`
	NodeID      uint64     `json:"nodeId"`
	NodeName    string     `json:"nodeName"`
	Status      string     `json:"status"`
	StartTime   *time.Time `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
	LogPath     string     `json:"logPath"`
	ErrorMsg    string     `json:"errorMsg"`
}

// TaskTemplateVO 模板视图对象
type TaskTemplateVO struct {
	ID             uint64     `json:"id"`
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	ConfigTemplate string     `json:"configTemplate"`
	CreatedAt      *time.Time `json:"createdAt"`
}
