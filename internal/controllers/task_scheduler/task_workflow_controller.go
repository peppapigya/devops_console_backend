package task_scheduler

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/request"
	tsSvc "devops-console-backend/internal/services/task_scheduler"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type TaskWorkflowController struct {
	workflowService *tsSvc.TaskWorkflowService
}

var errorMessages = map[string]string{
	"record not found":        "工作流不存在",
	"cron expression invalid": "定时表达式格式错误",
	"database error":          "系统繁忙，请稍后重试",
}

func handleError(err error) string {
	errStr := err.Error()
	for key, msg := range errorMessages {
		if strings.Contains(errStr, key) {
			return msg
		}
	}
	return "操作失败"
}

func NewTaskWorkflowController(service *tsSvc.TaskWorkflowService) *TaskWorkflowController {
	return &TaskWorkflowController{workflowService: service}
}

func (ctrl *TaskWorkflowController) List(c *gin.Context) {
	var req request.TaskWorkflowListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.BadRequest)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	status := -1
	if req.Status != nil {
		status = int(*req.Status)
	}
	list, err := ctrl.workflowService.ListWorkflows(req.Page, req.PageSize, req.Name, status)
	if err != nil {
		common.FailWithMsg(c, handleError(err))
		return
	}
	common.Success(c, list)
}

func (ctrl *TaskWorkflowController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, common.BadRequest)
		return
	}
	vo, err := ctrl.workflowService.GetWorkflowByID(id)
	if err != nil {
		common.FailWithMsg(c, handleError(err))
		return
	}
	common.Success(c, vo)
}

type combinedWorkflowRequest struct {
	request.TaskWorkflowCreateRequest
	Nodes []*request.TaskNodeCreateRequest `json:"nodes"`
	Edges []*request.TaskEdgeCreateRequest `json:"edges"`
}

type combinedWorkflowUpdateReq struct {
	request.TaskWorkflowUpdateRequest
	Nodes []*request.TaskNodeCreateRequest `json:"nodes"`
	Edges []*request.TaskEdgeCreateRequest `json:"edges"`
}

func (ctrl *TaskWorkflowController) Create(c *gin.Context) {
	var req combinedWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest)
		return
	}

	workflow, err := ctrl.workflowService.CreateWorkflow(&req.TaskWorkflowCreateRequest, req.Nodes, req.Edges)
	if err != nil {
		log.Printf("创建失败，失败原因：%v", err.Error())
		common.FailWithMsg(c, handleError(err))
		return
	}
	common.Success(c, gin.H{"id": workflow.ID})
}

func (ctrl *TaskWorkflowController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, common.BadRequest)
		return
	}
	var req combinedWorkflowUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest)
		return
	}
	if err := ctrl.workflowService.UpdateWorkflow(id, &req.TaskWorkflowUpdateRequest, req.Nodes, req.Edges); err != nil {
		common.FailWithMsg(c, handleError(err))
		return
	}
	common.Success(c, nil)
}

func (ctrl *TaskWorkflowController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, common.BadRequest)
		return
	}
	if err := ctrl.workflowService.DeleteWorkflow(id); err != nil {
		common.FailWithMsg(c, handleError(err))
		return
	}
	common.Success(c, nil)
}

func (ctrl *TaskWorkflowController) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, common.BadRequest)
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest)
		return
	}
	if err := ctrl.workflowService.UpdateWorkflowStatus(id, req.Status); err != nil {
		common.FailWithMsg(c, handleError(err))
		return
	}
	common.Success(c, nil)
}

func (ctrl *TaskWorkflowController) Execute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		common.FailWithMsg(c, "无效的工作流ID")
		return
	}

	userID := uint64(0)
	if uid, exists := c.Get("userID"); exists {
		userID = uid.(uint64)
	}

	executionID, err := ctrl.workflowService.ExecuteWorkflow(id, userID)
	if err != nil {
		common.FailWithMsg(c, err.Error())
		return
	}

	common.Success(c, gin.H{
		"execution_id": executionID,
		"message":      "工作流已开始执行",
	})
}
