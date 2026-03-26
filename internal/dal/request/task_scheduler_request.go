package request

// TaskWorkflowCreateRequest 创建工作流请求
type TaskWorkflowCreateRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	CronExpression string `json:"cronExpression"` // Cron expression
	Status         int8   `json:"status"`         // 1: enabled, 0: disabled
}

// TaskWorkflowUpdateRequest 更新工作流请求
type TaskWorkflowUpdateRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	CronExpression string `json:"cronExpression"`
	Status         int8   `json:"status"`
}

// TaskWorkflowListRequest 列表查询请求
type TaskWorkflowListRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
	Name     string `form:"name"`
	Status   *int8  `form:"status"`
}

// TaskNodeCreateRequest 创建节点请求
type TaskNodeCreateRequest struct {
	ID         uint64 `json:"id"`     // Database ID (if exists)
	TempID     string `json:"tempId"` // Temporary ID from frontend for mapping edges
	WorkflowID uint64 `json:"workflowId" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Type       string `json:"type" binding:"required"` // e.g., shell, python, http
	TargetID   uint64 `json:"targetId"`                // 目标ID（主机ID/K8s实例ID/数据库实例ID）
	TargetType string `json:"targetType"`              // 目标类型：host/k8s/db
	Config     string `json:"config"`                  // JSON config for node execution
	PositionX  int    `json:"position_x"`              // 画布坐标X
	PositionY  int    `json:"position_y"`              // 画布坐标Y
}

// TaskEdgeCreateRequest 创建边请求
type TaskEdgeCreateRequest struct {
	WorkflowID   uint64 `json:"workflowId" binding:"required"`
	SourceTempID string `json:"sourceTempId"`
	TargetTempID string `json:"targetTempId"`
	FromNodeID   uint64 `json:"fromNodeId"`
	ToNodeID     uint64 `json:"toNodeId"`
	SourceHandle string `json:"sourceHandle"` // 源节点连接点：top/bottom/left/right
	TargetHandle string `json:"targetHandle"` // 目标节点连接点：top/bottom/left/right
	Condition    string `json:"condition"`    // Condition for following this edge
}

// TaskExecutionListRequest 执行记录列表请求
type TaskExecutionListRequest struct {
	Page       int     `form:"page" binding:"required,min=1"`
	PageSize   int     `form:"pageSize" binding:"required,min=1,max=100"`
	WorkflowID *uint64 `form:"workflowId"`
	Status     *string `form:"status"`
	StartTime  *int64  `form:"startTime"`
	EndTime    *int64  `form:"endTime"`
}

// ManualTriggerRequest 手动触发请求
type ManualTriggerRequest struct {
	WorkflowID uint64            `json:"workflowId" binding:"required"`
	Params     map[string]string `json:"params"`
}
