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

// SysPositionController 岗位管理控制器
type SysPositionController struct {
	posMapper *mapper.SysPositionMapper
}

func NewSysPositionController(pm *mapper.SysPositionMapper) *SysPositionController {
	return &SysPositionController{posMapper: pm}
}

// List 分页查询岗位
func (c *SysPositionController) List(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqSys.PositionPageRequest
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
	total, list, err := c.posMapper.ListPage(req.Page, req.PageSize, req.Name, req.Code, req.Status)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	vos := make([]response.PositionVO, 0, len(list))
	for _, p := range list {
		vo := response.PositionVO{ID: p.ID, Name: p.Name, Code: p.Code, Sort: p.Sort, Status: p.Status, CreatedAt: p.CreatedAt}
		if p.Remark != nil {
			vo.Remark = *p.Remark
		}
		vos = append(vos, vo)
	}
	helper.SuccessWithData("查询成功", "data", response.PositionPageVO{Total: total, List: vos})
}

// ListAll 查询全部岗位（供下拉）
func (c *SysPositionController) ListAll(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	list, err := c.posMapper.ListAll()
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("查询成功", "data", list)
}

// Create 新建岗位
func (c *SysPositionController) Create(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqSys.PositionCreateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	exists, _ := c.posMapper.ExistsByCode(req.Code, 0)
	if exists {
		helper.Fail(common.NewErrorCode(400, "岗位编码已存在"))
		return
	}
	now := time.Now()
	pos := &model.SysPosition{Name: req.Name, Code: req.Code, Sort: req.Sort, Status: req.Status}
	if req.Remark != "" {
		pos.Remark = &req.Remark
	}
	pos.CreatedAt = &now
	pos.UpdatedAt = &now

	if err := c.posMapper.Create(pos); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("创建成功", "data", gin.H{"id": pos.ID})
}

// Update 更新岗位
func (c *SysPositionController) Update(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqSys.PositionUpdateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	exists, _ := c.posMapper.ExistsByCode(req.Code, id)
	if exists {
		helper.Fail(common.NewErrorCode(400, "岗位编码已存在"))
		return
	}
	pos, err := c.posMapper.GetByID(id)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	pos.Name = req.Name
	pos.Code = req.Code
	pos.Sort = req.Sort
	pos.Status = req.Status
	if req.Remark != "" {
		pos.Remark = &req.Remark
	} else {
		pos.Remark = nil
	}

	if err := c.posMapper.Update(pos); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("更新成功", "data", nil)
}

// Delete 删除岗位
func (c *SysPositionController) Delete(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.posMapper.SoftDelete(id); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("删除成功", "data", nil)
}
