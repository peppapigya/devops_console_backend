package cicd

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type ProjectsController struct {
	mapper *mapper.ProjectMapper
}

func NewProjectsController(mapper *mapper.ProjectMapper) *ProjectsController {
	return &ProjectsController{
		mapper: mapper,
	}
}
func (c *ProjectsController) GetProjectById(ctx *gin.Context) {
	var id uint32
	utils.GetParam(ctx, "id", &id, nil)
	helper := utils.NewResponseHelper(ctx)
	project, err := c.mapper.GetProjectById(id)
	if err != nil {
		helper.Error(500, "查询数据失败")
		return
	}
	helper.SuccessWithData("成功", "data", project)
}
func (c *ProjectsController) UpdateProject(ctx *gin.Context) {
	var project *model.Project
	helper := utils.NewResponseHelper(ctx)
	if ok := utils.BindAndValidate(ctx, &project); !ok {
		helper.ValidationError("参数验证失败")
		return
	}
	if err := c.mapper.UpdateProject(project); err != nil {
		helper.InternalError("更新数据失败")
	} else {
		helper.Success("更新数据成功")
	}
}
func (c *ProjectsController) CreateProject(ctx *gin.Context) {
	var project *model.Project
	helper := utils.NewResponseHelper(ctx)
	if ok := utils.BindAndValidate(ctx, &project); !ok {
		helper.ValidationError("参数验证失败")
	}
	if err := c.mapper.CreateProject(project); err != nil {
		helper.InternalError("创建数据失败")
	} else {
		helper.Success("创建数据成功")
	}
}
func (c *ProjectsController) DeleteProject(ctx *gin.Context) {
	var id uint32
	helper := utils.NewResponseHelper(ctx)
	utils.GetParam(ctx, "id", &id, nil)
	if err := c.mapper.DeleteProject(id); err != nil {
		helper.InternalError("删除数据失败")
	} else {
		helper.Success("删除数据成功")
	}
}
func (c *ProjectsController) GetPageProjects(ctx *gin.Context) {
	var pageNum int
	var pageSize int
	helper := utils.NewResponseHelper(ctx)
	utils.GetParam(ctx, "pageNum", &pageNum, nil)
	utils.GetParam(ctx, "pageSize", &pageSize, nil)
	projects, total, err := c.mapper.GetPageProjects(pageNum, pageSize)
	response := common.PageInfoResponse[*model.Project]{
		Data:     projects,
		PageNum:  pageNum,
		PageSize: pageSize,
		Total:    total,
	}
	if err != nil {
		helper.InternalError("查询数据失败")
	} else {
		helper.SuccessWithData("成功", "data", response)
	}
}
func (c *ProjectsController) GetProjects(ctx *gin.Context) {
	var name string
	helper := utils.NewResponseHelper(ctx)
	utils.GetParam(ctx, "name", &name, nil)
	projects, err := c.mapper.GetProjectsByName(name)
	if err != nil {
		helper.InternalError("查询数据失败")
	} else {
		helper.SuccessWithData("成功", "data", projects)
	}
}
