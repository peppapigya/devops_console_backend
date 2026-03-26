package task_scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/services/task_scheduler/executor"
	"devops-console-backend/internal/websocket"
)

// DAGExecutor DAG执行器
type DAGExecutor struct {
	executionMapper *mapper.TaskExecutionMapper
	nodeExecMapper  *mapper.TaskNodeExecutionMapper
}

func NewDAGExecutor(executionMapper *mapper.TaskExecutionMapper, nodeExecMapper *mapper.TaskNodeExecutionMapper) *DAGExecutor {
	return &DAGExecutor{
		executionMapper: executionMapper,
		nodeExecMapper:  nodeExecMapper,
	}
}

// ExecuteWorkflow 执行工作流（手动触发）
func (d *DAGExecutor) ExecuteWorkflow(ctx context.Context, workflowID uint64, nodes []*model.TaskNode, edges []*model.TaskEdge, triggeredBy uint64) (uint64, error) {
	return d.ExecuteWorkflowWithTrigger(ctx, workflowID, nodes, edges, triggeredBy, "manual")
}

// ExecuteWorkflowWithTrigger 执行工作流（指定触发类型）
func (d *DAGExecutor) ExecuteWorkflowWithTrigger(ctx context.Context, workflowID uint64, nodes []*model.TaskNode, edges []*model.TaskEdge, triggeredBy uint64, triggerType string) (uint64, error) {
	// 1. 创建执行记录
	StartTime := time.Now()
	execution := &model.TaskExecution{
		WorkflowID:  uint32(workflowID),
		ExecutionNo: generateExecutionNo(),
		TriggerType: triggerType,
		Status:      "running",
		StartTime:   &StartTime,
		TriggeredBy: uint32(triggeredBy),
	}

	if err := d.executionMapper.Create(execution); err != nil {
		return 0, err
	}

	// 2. 异步执行DAG
	go d.runDAG(ctx, uint64(execution.ID), nodes, edges)

	return uint64(execution.ID), nil
}

// runDAG 运行DAG
func (d *DAGExecutor) runDAG(ctx context.Context, executionID uint64, nodes []*model.TaskNode, edges []*model.TaskEdge) {
	// 构建依赖图
	dependencies := make(map[uint64][]uint64)
	inDegree := make(map[uint64]int)

	for _, node := range nodes {
		inDegree[uint64(node.ID)] = 0
	}

	for _, edge := range edges {
		dependencies[uint64(edge.SourceNodeID)] = append(dependencies[uint64(edge.SourceNodeID)], uint64(edge.TargetNodeID))
		inDegree[uint64(edge.TargetNodeID)]++
	}

	// 拓扑排序执行
	for len(inDegree) > 0 {
		var readyNodes []uint64
		for nodeID, degree := range inDegree {
			if degree == 0 {
				readyNodes = append(readyNodes, nodeID)
			}
		}

		if len(readyNodes) == 0 {
			d.updateExecutionStatus(executionID, "failed", "工作流存在循环依赖")
			return
		}

		// 执行就绪节点
		for _, nodeID := range readyNodes {
			node := findNodeByID(nodes, nodeID)
			if node != nil {
				d.executeNode(ctx, executionID, node)
			}
			delete(inDegree, nodeID)

			for _, nextID := range dependencies[nodeID] {
				inDegree[nextID]--
			}
		}
	}

	d.updateExecutionStatus(executionID, "success", "")
}

func (d *DAGExecutor) executeNode(ctx context.Context, executionID uint64, node *model.TaskNode) {
	// 创建节点执行记录
	startTime := time.Now()
	nodeExecution := &model.TaskNodeExecution{
		ExecutionID: uint32(executionID),
		NodeID:      node.ID,
		Status:      "running",
		StartTime:   &startTime,
	}
	d.nodeExecMapper.Create(nodeExecution)

	logger := &SimpleLogger{executionID: executionID}
	logger.Log("info", fmt.Sprintf(">>> 开始执行节点: %s (ID: %d, 类型: %s)", node.NodeName, node.ID, node.NodeType))

	// 获取执行器
	factory := executor.GetExecutorFactory()
	exec, ok := factory.GetExecutor(node.NodeType)
	if !ok {
		result := &executor.ExecutionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("不支持的任务类型: %s", node.NodeType),
		}
		logger.Log("error", fmt.Sprintf("不支持的任务类型: %s", node.NodeType))
		d.updateNodeExecution(uint64(nodeExecution.ID), result, logger)
		return
	}

	// 解析配置
	var config map[string]interface{}
	if node.Config != nil && *node.Config != "" {
		if err := json.Unmarshal([]byte(*node.Config), &config); err != nil {
			result := &executor.ExecutionResult{
				Success:  false,
				ErrorMsg: fmt.Sprintf("解析节点配置失败: %v", err),
			}
			logger.Log("error", fmt.Sprintf("解析节点配置失败: %v", err))
			d.updateNodeExecution(uint64(nodeExecution.ID), result, logger)
			return
		}
	}

	// 构建执行上下文
	execCtx := &executor.TaskExecutionContext{
		ExecutionID: executionID,
		NodeID:      uint64(node.ID),
		Config:      config,
		TargetID:    node.TargetID,
		TargetType:  node.TargetType,
		Logger:      logger,
	}

	// 执行
	result := exec.Execute(ctx, execCtx)

	// 更新结果
	if result.Success {
		logger.Log("info", fmt.Sprintf("<<< 节点执行成功: %s (耗时: %dms)", node.NodeName, result.Duration))
	} else {
		logger.Log("error", fmt.Sprintf("<<< 节点执行失败: %s (原因: %s)", node.NodeName, result.ErrorMsg))
	}

	d.updateNodeExecution(uint64(nodeExecution.ID), result, logger)
}

func (d *DAGExecutor) updateExecutionStatus(executionID uint64, status string, errorMsg string) {
	execution, err := d.executionMapper.GetByID(executionID)
	if err != nil || execution == nil {
		fmt.Printf("Error getting execution: %v\n", err)
		return
	}

	endTime := time.Now()
	execution.Status = status
	execution.EndTime = &endTime

	if execution.StartTime != nil {
		execution.Duration = int32(endTime.Sub(*execution.StartTime).Milliseconds())
	}

	if errorMsg != "" {
		execution.ErrorMessage = &errorMsg
	}
	d.executionMapper.Update(execution)
}

func (d *DAGExecutor) updateNodeExecution(nodeExecutionID uint64, result *executor.ExecutionResult, logger *SimpleLogger) {
	status := "success"
	if !result.Success {
		status = "failed"
	}

	outputParams := string(mustJSON(result.OutputVars))
	endTime := time.Now()
	nodeExec := &model.TaskNodeExecution{
		ID:           uint32(nodeExecutionID),
		Status:       status,
		EndTime:      &endTime,
		Duration:     int32(result.Duration),
		OutputParams: &outputParams,
		ErrorMessage: &result.ErrorMsg,
	}

	logs := logger.GetLogs()
	if logs != "" {
		nodeExec.Logs = &logs
	}

	d.nodeExecMapper.Update(nodeExec)
}

func findNodeByID(nodes []*model.TaskNode, id uint64) *model.TaskNode {
	for _, n := range nodes {
		if n.ID == uint32(id) {
			return n
		}
	}
	return nil
}

func generateExecutionNo() string {
	return fmt.Sprintf("EXEC-%d", time.Now().Unix())
}

func mustJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

// SimpleLogger 简单日志实现
type SimpleLogger struct {
	executionID uint64
	logs        []string
}

func (l *SimpleLogger) Log(level string, message string) {
	logEntry := fmt.Sprintf("[%s] %s: %s", time.Now().Format("2006-01-02 15:04:05"), level, message)
	l.logs = append(l.logs, logEntry)
	fmt.Println(logEntry)

	if l.executionID > 0 {
		websocket.BroadcastLog(l.executionID, logEntry)
	}
}

func (l *SimpleLogger) GetLogs() string {
	return strings.Join(l.logs, "\n")
}
