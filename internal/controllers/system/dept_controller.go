package system

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	reqSys "devops-console-backend/internal/dal/request/system"
	"devops-console-backend/pkg/utils"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// SysDeptController 部门管理控制器
type SysDeptController struct {
	deptMapper *mapper.SysDeptMapper
}

func NewSysDeptController(dm *mapper.SysDeptMapper) *SysDeptController {
	return &SysDeptController{deptMapper: dm}
}

// List 查询部门树
func (c *SysDeptController) List(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	name := ctx.Query("name")
	var status *uint8
	if s := ctx.Query("status"); s != "" {
		var sv uint8
		if _, err := fmt.Sscanf(s, "%d", &sv); err == nil {
			status = &sv
		}
	}
	depts, err := c.deptMapper.ListAll(name, status)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	tree := mapper.BuildDeptTree(depts, 0)
	helper.SuccessWithData("查询成功", "data", tree)
}

// Create 新建部门
func (c *SysDeptController) Create(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqSys.DeptCreateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	now := time.Now()
	dept := &model.SysDepartment{
		ParentID: req.ParentID,
		Name:     req.Name,
		Sort:     req.Sort,
		Status:   req.Status,
	}
	if req.Leader != "" {
		dept.Leader = &req.Leader
	}
	if req.Phone != "" {
		dept.Phone = &req.Phone
	}
	if req.Email != "" {
		dept.Email = &req.Email
	}
	if req.Remark != "" {
		dept.Remark = &req.Remark
	}
	dept.CreatedAt = &now
	dept.UpdatedAt = &now

	if err := c.deptMapper.Create(dept); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("创建成功", "data", gin.H{"id": dept.ID})
}

// Update 更新部门
func (c *SysDeptController) Update(ctx *gin.Context) {

	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqSys.DeptUpdateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	dept, err := c.deptMapper.GetByID(id)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	dept.ParentID = req.ParentID
	dept.Name = req.Name
	dept.Sort = req.Sort
	dept.Status = req.Status

	if req.Leader != "" {
		dept.Leader = &req.Leader
	} else {
		dept.Leader = nil
	}
	if req.Phone != "" {
		dept.Phone = &req.Phone
	} else {
		dept.Phone = nil
	}
	if req.Email != "" {
		dept.Email = &req.Email
	} else {
		dept.Email = nil
	}
	if req.Remark != "" {
		dept.Remark = &req.Remark
	} else {
		dept.Remark = nil
	}

	if err := c.deptMapper.Update(dept); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("更新成功", "data", nil)
}

// Delete 删除部门
func (c *SysDeptController) Delete(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.deptMapper.SoftDelete(id); err != nil {
		helper.Fail(common.NewErrorCode(400, err.Error()))
		return
	}
	helper.SuccessWithData("删除成功", "data", nil)
}
