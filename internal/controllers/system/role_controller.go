package system

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	reqSys "devops-console-backend/internal/dal/request/system"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// SysRoleController 角色管理控制器
type SysRoleController struct {
	roleMapper *mapper.SysRoleMapper
	menuMapper *mapper.SysMenuMapper
}

func NewSysRoleController(rm *mapper.SysRoleMapper, mm *mapper.SysMenuMapper) *SysRoleController {
	return &SysRoleController{roleMapper: rm, menuMapper: mm}
}

// List 分页查询角色
func (c *SysRoleController) List(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqSys.RolePageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	total, list, err := c.roleMapper.ListPage(req.Page, req.PageSize, req.Name, req.Code, req.Status)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	vos := make([]response.RoleVO, 0, len(list))
	for _, r := range list {
		vo := response.RoleVO{ID: r.ID, Name: r.Name, Code: r.Code, Sort: r.Sort, Status: r.Status, CreatedAt: r.CreatedAt}
		if r.Remark != nil {
			vo.Remark = *r.Remark
		}
		vos = append(vos, vo)
	}
	helper.SuccessWithData("查询成功", "data", response.RolePageVO{Total: total, List: vos})
}

// ListAll 全部角色（下拉）
func (c *SysRoleController) ListAll(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	list, err := c.roleMapper.ListAll()
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", list)
}

// Create 新建角色
func (c *SysRoleController) Create(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqSys.RoleCreateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	exists, _ := c.roleMapper.ExistsByCode(req.Code, 0)
	if exists {
		helper.Fail(common.NewErrorCode(400, "角色编码已存在"))
		return
	}
	now := time.Now()
	role := &model.SysRole{Name: req.Name, Code: req.Code, Sort: req.Sort, Status: req.Status}
	if req.Remark != "" {
		role.Remark = &req.Remark
	}
	role.CreatedAt = &now
	role.UpdatedAt = &now

	if err := c.roleMapper.Create(role); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("创建成功", "data", gin.H{"id": role.ID})
}

// Update 更新角色
func (c *SysRoleController) Update(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqSys.RoleUpdateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	exists, _ := c.roleMapper.ExistsByCode(req.Code, id)
	if exists {
		helper.Fail(common.NewErrorCode(400, "角色编码已存在"))
		return
	}
	role, err := c.roleMapper.GetByID(id)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	role.Name = req.Name
	role.Code = req.Code
	role.Sort = req.Sort
	role.Status = req.Status
	if req.Remark != "" {
		role.Remark = &req.Remark
	} else {
		role.Remark = nil
	}

	if err := c.roleMapper.Update(role); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("更新成功", "data", nil)
}

// Delete 删除角色
func (c *SysRoleController) Delete(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.roleMapper.SoftDelete(id); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("删除成功", "data", nil)
}

// GetMenuIDs 获取角色已分配的菜单 ID 列表
func (c *SysRoleController) GetMenuIDs(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	ids, err := c.roleMapper.GetRoleMenuIDs(id)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", ids)
}

// AssignMenus 给角色分配菜单
func (c *SysRoleController) AssignMenus(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqSys.RoleMenuAssignRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.roleMapper.AssignMenus(id, req.MenuIDs); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("权限分配成功", "data", nil)
}
