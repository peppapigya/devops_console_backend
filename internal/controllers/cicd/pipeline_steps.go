package cicd

import (
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type PipelineStepsController struct {
	mapper *mapper.PipelineStepsMapper
}

func NewPipelineStepsController(mapper *mapper.PipelineStepsMapper) *PipelineStepsController {
	return &PipelineStepsController{
		mapper: mapper,
	}
}

func (c *PipelineStepsController) GetPipelineStepById(ctx *gin.Context) {
	var id uint32
	utils.GetParam(ctx, "id", &id, nil)
	helper := utils.NewResponseHelper(ctx)
	pipelineStep, err := c.mapper.GetPipelineStepById(id)
	if err != nil {
		helper.DatabaseError("获取流水线步骤详情失败")
		return
	}
	helper.SuccessWithData("success", "pipelineStep", pipelineStep)
}

func (c *PipelineStepsController) UpdatePipelineStep(ctx *gin.Context) {
	pipelineStep := &model.PipelineStep{}
	utils.BindAndValidate(ctx, pipelineStep)
	helper := utils.NewResponseHelper(ctx)
	err := c.mapper.UpdatePipelineStep(pipelineStep)
	if err != nil {
		helper.DatabaseError("更新流水线步骤详情失败")
		return
	}
	helper.Success("success")
}

func (c *PipelineStepsController) CreatePipelineStep(ctx *gin.Context) {
	pipelineStep := &model.PipelineStep{}
	utils.BindAndValidate(ctx, pipelineStep)
	helper := utils.NewResponseHelper(ctx)
	err := c.mapper.CreatePipelineStep(pipelineStep)
	if err != nil {
		helper.DatabaseError("创建流水线步骤详情失败")
		return
	}
	helper.SuccessWithData("success", "data", pipelineStep)
}

func (c *PipelineStepsController) GetPipelineSteps(ctx *gin.Context) {
	var pipelineId uint32
	utils.GetParam(ctx, "pipelineId", &pipelineId, nil)
	helper := utils.NewResponseHelper(ctx)
	steps, err := c.mapper.GetPipelineStepByPipelineId(pipelineId)
	if err != nil {
		helper.DatabaseError("获取流水线步骤列表失败")
		return
	}
	helper.SuccessWithData("success", "data", steps)
}

func (c *PipelineStepsController) DeletePipelineStep(ctx *gin.Context) {
	var id uint32
	utils.GetParam(ctx, "id", &id, nil)
	helper := utils.NewResponseHelper(ctx)
	if err := c.mapper.DeletePipelineStep(id); err != nil {
		helper.DatabaseError("删除流水线步骤失败")
		return
	}
	helper.Success("删除成功")
}
