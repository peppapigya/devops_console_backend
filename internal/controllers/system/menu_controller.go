package system

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	reqSys "devops-console-backend/internal/dal/request/system"
	"devops-console-backend/pkg/utils"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysMenuController 菜单管理控制器
type SysMenuController struct {
	menuMapper *mapper.SysMenuMapper
}

func NewSysMenuController(mm *mapper.SysMenuMapper) *SysMenuController {
	return &SysMenuController{menuMapper: mm}
}

// List 查询所有菜单（树形）
func (c *SysMenuController) List(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	menus, err := c.menuMapper.ListAll()
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	tree := mapper.BuildMenuTree(menus, 0)
	helper.SuccessWithData("查询成功", "data", tree)
}

// Create 新建菜单
func (c *SysMenuController) Create(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqSys.MenuCreateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	now := time.Now()
	menu := &model.SysMenu{
		ParentID: req.ParentID,
		Name:     req.Name,
		Type:     req.Type,
		Sort:     req.Sort,
		Visible:  req.Visible,
		Status:   req.Status,
	}
	if req.Path != "" {
		menu.Path = &req.Path
	}
	if req.Component != "" {
		menu.Component = &req.Component
	}
	if req.Icon != "" {
		menu.Icon = &req.Icon
	}
	if req.Perm != "" {
		menu.Perm = &req.Perm
	}
	menu.CreatedAt = &now
	menu.UpdatedAt = &now

	if err := c.menuMapper.Create(menu); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("创建成功", "data", gin.H{"id": menu.ID})
}

// Update 更新菜单
func (c *SysMenuController) Update(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqSys.MenuUpdateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	menu, err := c.menuMapper.GetByID(id)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	menu.ParentID = req.ParentID
	menu.Name = req.Name
	menu.Type = req.Type
	menu.Sort = req.Sort
	menu.Visible = req.Visible
	menu.Status = req.Status
	if req.Path != "" {
		menu.Path = &req.Path
	} else {
		menu.Path = nil
	}
	if req.Component != "" {
		menu.Component = &req.Component
	} else {
		menu.Component = nil
	}
	if req.Icon != "" {
		menu.Icon = &req.Icon
	} else {
		menu.Icon = nil
	}
	if req.Perm != "" {
		menu.Perm = &req.Perm
	} else {
		menu.Perm = nil
	}

	if err := c.menuMapper.Update(menu); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("更新成功", "data", nil)
}

// Delete 删除菜单
func (c *SysMenuController) Delete(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.menuMapper.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.Fail(common.NewErrorCode(400, "该菜单下存在子菜单，请先删除子菜单"))
		} else {
			helper.InternalError(err.Error())
		}
		return
	}
	helper.SuccessWithData("删除成功", "data", nil)
}
