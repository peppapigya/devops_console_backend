package task_scheduler

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskExecutionController struct {
	executionMapper *mapper.TaskExecutionMapper
	nodeExecMapper  *mapper.TaskNodeExecutionMapper
}

func NewTaskExecutionController(executionMapper *mapper.TaskExecutionMapper, nodeExecMapper *mapper.TaskNodeExecutionMapper) *TaskExecutionController {
	return &TaskExecutionController{
		executionMapper: executionMapper,
		nodeExecMapper:  nodeExecMapper,
	}
}

// ListExecutions 查询执行历史
func (c *TaskExecutionController) ListExecutions(ctx *gin.Context) {
	workflowID := ctx.Query("workflowId")
	status := ctx.Query("status")
	page := ctx.DefaultQuery("page", "1")
	pageSize := ctx.DefaultQuery("pageSize", "10")

	// 转换参数
	wID, _ := strconv.ParseUint(workflowID, 10, 64)
	p, _ := strconv.Atoi(page)
	ps, _ := strconv.Atoi(pageSize)

	// 查询
	total, list, err := c.executionMapper.ListPage(p, ps, wID, status, "")
	if err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}

	common.Success(ctx, gin.H{
		"total":    total,
		"list":     list,
		"page":     p,
		"pageSize": ps,
	})
}

// GetExecutionDetail 获取执行详情（包含节点日志）
func (c *TaskExecutionController) GetExecutionDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		common.FailWithMsg(ctx, "无效的执行ID")
		return
	}

	execution, err := c.executionMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "执行记录查询失败: "+err.Error())
		return
	}
	if execution == nil {
		common.FailWithMsg(ctx, "执行记录不存在")
		return
	}

	nodes, err := c.nodeExecMapper.ListByExecutionID(id)
	if err != nil {
		common.FailWithMsg(ctx, "获取节点日志失败: "+err.Error())
		return
	}

	common.Success(ctx, gin.H{
		"execution": execution,
		"nodes":     nodes,
	})
}

func (c *TaskExecutionController) GetExecutionLogs(ctx *gin.Context) {
	executionIDStr := ctx.Param("id")
	executionID, err := strconv.ParseUint(executionIDStr, 10, 64)
	if err != nil {
		common.FailWithMsg(ctx, "无效的执行ID")
		return
	}

	execution, err := c.executionMapper.GetByID(executionID)
	if err != nil {
		common.FailWithMsg(ctx, "获取执行记录失败")
		return
	}
	if execution == nil {
		common.FailWithMsg(ctx, "执行记录不存在")
		return
	}

	nodeExecutions, err := c.nodeExecMapper.ListByExecutionID(executionID)
	if err != nil {
		common.FailWithMsg(ctx, "获取节点日志失败")
		return
	}

	common.Success(ctx, gin.H{
		"execution": execution,
		"nodes":     nodeExecutions,
	})
}
