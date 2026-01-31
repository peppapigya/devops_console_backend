package cicd

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type PipelinesController struct {
	mapper *mapper.PipelinesMapper
}

func NewPipelinesController(mapper *mapper.PipelinesMapper) *PipelinesController {
	return &PipelinesController{
		mapper: mapper,
	}
}
func (c *PipelinesController) GetPipelineById(ctx *gin.Context) {
	var id uint32
	utils.GetParam(ctx, "id", &id, nil)
	helper := utils.NewResponseHelper(ctx)
	pipeline, err := c.mapper.GetPipelineById(id)
	if err != nil {
		helper.DatabaseError(err.Error())
		return
	}
	helper.SuccessWithData("成功", "data", pipeline)
}
func (c *PipelinesController) UpdatePipeline(ctx *gin.Context) {
	pipeline := &model.Pipeline{}
	helper := utils.NewResponseHelper(ctx)
	if !utils.BindAndValidate(ctx, pipeline) {
		helper.ValidationError("参数验证失败")
		return
	}
	if err := c.mapper.UpdatePipeline(pipeline); err != nil {
		helper.DatabaseError(err.Error())
		return
	}

	helper.Success("更新数据成功")
}

func (c *PipelinesController) CreatePipeline(ctx *gin.Context) {
	pipeline := &model.Pipeline{}
	helper := utils.NewResponseHelper(ctx)
	if !utils.BindAndValidate(ctx, pipeline) {
		helper.ValidationError("参数验证失败")
		return
	}
	if err := c.mapper.CreatePipeline(pipeline); err != nil {
		helper.DatabaseError(err.Error())
		return
	}

	helper.Success("创建数据成功")
}

func (c *PipelinesController) DeletePipeline(ctx *gin.Context) {
	var id uint32
	helper := utils.NewResponseHelper(ctx)
	utils.GetParam(ctx, "id", &id, nil)
	if err := c.mapper.DeletePipeline(id); err != nil {
		helper.DatabaseError(err.Error())
		return
	}

	helper.Success("删除数据成功")
}
func (c *PipelinesController) GetPagePipelines(ctx *gin.Context) {
	var pageNum int
	var pageSize int
	helper := utils.NewResponseHelper(ctx)
	utils.GetParam(ctx, "pageNum", &pageNum, nil)
	utils.GetParam(ctx, "pageSize", &pageSize, nil)
	pipelines, total, err := c.mapper.GetPagePipelines(pageNum, pageSize)
	if err != nil {
		helper.DatabaseError(err.Error())
		return
	}
	response := common.PageInfoResponse[*model.Pipeline]{
		Data:     pipelines,
		PageNum:  pageNum,
		PageSize: pageSize,
		Total:    total,
	}
	helper.SuccessWithData("成功", "data", response)
}
