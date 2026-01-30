package cicd

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type PipelineRunController struct {
	mapper *mapper.PipelineRunMapper
	helper *utils.ResponseHelper
}

func NewPipelineRunController(mapper *mapper.PipelineRunMapper, helper *utils.ResponseHelper) *PipelineRunController {
	return &PipelineRunController{
		mapper: mapper,
		helper: helper,
	}
}

func (c *PipelineRunController) GetPipelineRunById(ctx *gin.Context) {
	var id uint64
	utils.GetParam(ctx, "id", &id, nil)
	pipelineRun, err := c.mapper.GetPipelineRunById(id)
	if err != nil {
		c.helper.DatabaseError(err.Error())
	}
	c.helper.SuccessWithData("success", "pipelineRun", pipelineRun)
}

func (c *PipelineRunController) UpdatePipelineRun(ctx *gin.Context) {
	var pipelineRun model.PipelineRun
	if !utils.BindAndValidate(ctx, &pipelineRun) {
		return
	}
	err := c.mapper.UpdatePipelineRun(&pipelineRun)
	if err != nil {
		c.helper.DatabaseError(err.Error())
	}
	c.helper.Success("success")
}

func (c *PipelineRunController) CreatePipelineRun(ctx *gin.Context) {
	var pipelineRun model.PipelineRun
	if !utils.BindAndValidate(ctx, &pipelineRun) {
		return
	}
	err := c.mapper.CreatePipelineRun(&pipelineRun)
	if err != nil {
		c.helper.DatabaseError(err.Error())
	}
	c.helper.SuccessWithData("success", "data", pipelineRun)
}

func (c *PipelineRunController) DeletePipelineRun(ctx *gin.Context) {
	var id uint64
	utils.GetParam(ctx, "id", &id, nil)
	err := c.mapper.DeletePipelineRun(id)
	if err != nil {
		c.helper.DatabaseError(err.Error())
	}
	c.helper.Success("success")
}

func (c *PipelineRunController) GetPagePipelineRuns(ctx *gin.Context) {
	var pageNum int
	var pageSize int
	utils.GetParam(ctx, "pageNum", &pageNum, nil)
	utils.GetParam(ctx, "pageSize", &pageSize, nil)
	pipelineRuns, total, err := c.mapper.GetPagePipelineRuns(pageNum, pageSize)
	response := &common.PageInfoResponse[*model.PipelineRun]{
		Data:     pipelineRuns,
		PageNum:  pageNum,
		PageSize: pageSize,
		Total:    total,
	}
	if err != nil {
		c.helper.DatabaseError(err.Error())
	}
	c.helper.SuccessWithData("success", "data", response)
}
