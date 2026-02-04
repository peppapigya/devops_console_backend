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
}

func NewPipelineRunController(mapper *mapper.PipelineRunMapper) *PipelineRunController {
	return &PipelineRunController{
		mapper: mapper,
	}
}

func (c *PipelineRunController) GetPipelineRunById(ctx *gin.Context) {
	var id uint64
	helper := utils.NewResponseHelper(ctx)
	utils.GetParam(ctx, "id", &id, nil)
	pipelineRun, err := c.mapper.GetPipelineRunById(id)
	if err != nil {
		helper.DatabaseError(err.Error())
	}
	helper.SuccessWithData("success", "pipelineRun", pipelineRun)
}

func (c *PipelineRunController) UpdatePipelineRun(ctx *gin.Context) {
	var pipelineRun model.PipelineRun
	helper := utils.NewResponseHelper(ctx)
	if !utils.BindAndValidate(ctx, &pipelineRun) {
		return
	}
	err := c.mapper.UpdatePipelineRun(&pipelineRun)
	if err != nil {
		helper.DatabaseError(err.Error())
	}
	helper.Success("success")
}

func (c *PipelineRunController) CreatePipelineRun(ctx *gin.Context) {
	var pipelineRun model.PipelineRun
	helper := utils.NewResponseHelper(ctx)
	if !utils.BindAndValidate(ctx, &pipelineRun) {
		return
	}
	err := c.mapper.CreatePipelineRun(&pipelineRun)
	if err != nil {
		helper.DatabaseError(err.Error())
	}
	helper.SuccessWithData("success", "data", pipelineRun)
}

func (c *PipelineRunController) DeletePipelineRun(ctx *gin.Context) {
	var id uint64
	helper := utils.NewResponseHelper(ctx)
	utils.GetParam(ctx, "id", &id, nil)
	err := c.mapper.DeletePipelineRun(id)
	if err != nil {
		helper.DatabaseError(err.Error())
	}
	helper.Success("success")
}

func (c *PipelineRunController) GetPagePipelineRuns(ctx *gin.Context) {
	var pageNum int
	var pageSize int
	helper := utils.NewResponseHelper(ctx)
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
		helper.DatabaseError(err.Error())
	}
	helper.SuccessWithData("success", "data", response)
}
