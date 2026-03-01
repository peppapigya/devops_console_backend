package asset

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	reqAsset "devops-console-backend/internal/dal/request/asset"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// HostController 主机管理控制器
type HostController struct {
	groupMapper *mapper.AssetHostGroupMapper
	hostMapper  *mapper.AssetHostMapper
}

func NewHostController(gm *mapper.AssetHostGroupMapper, hm *mapper.AssetHostMapper) *HostController {
	return &HostController{groupMapper: gm, hostMapper: hm}
}

// ==================== 分组接口 ====================

// ListGroups 查询所有分组（树形）
func (c *HostController) ListGroups(ctx *gin.Context) {
	groups, err := c.groupMapper.ListAll()
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	// 附加每组主机数
	for i := range groups {
		var cnt int64
		c.hostMapper.DB.Model(&model.AssetHost{}).
			Where("group_id = ? AND deleted_at IS NULL", groups[i].ID).Count(&cnt)
		groups[i].HostCount = cnt
	}
	tree := mapper.BuildHostGroupTree(groups, 0)
	common.Success(ctx, gin.H{"data": tree})
}

// CreateGroup 新建分组
func (c *HostController) CreateGroup(ctx *gin.Context) {
	var req reqAsset.HostGroupCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	now := time.Now()
	remark := req.Remark
	g := &model.AssetHostGroup{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Remark:    &remark,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
	if err := c.groupMapper.Create(g); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": g.ID}})
}

// UpdateGroup 编辑分组
func (c *HostController) UpdateGroup(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqAsset.HostGroupUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	g, err := c.groupMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "分组不存在")
		return
	}
	now := time.Now()
	g.ParentID = req.ParentID
	g.Name = req.Name
	g.Remark = &req.Remark
	g.UpdatedAt = &now
	if err := c.groupMapper.Update(g); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// DeleteGroup 删除分组
func (c *HostController) DeleteGroup(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.groupMapper.SoftDelete(id); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// ==================== 主机接口 ====================

// ListHosts 分页查询主机
func (c *HostController) ListHosts(ctx *gin.Context) {
	var req reqAsset.HostPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	total, list, err := c.hostMapper.ListPage(req.Page, req.PageSize, req.GroupID, req.Name, req.IP, req.Status)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

// CreateHost 新建主机
func (c *HostController) CreateHost(ctx *gin.Context) {
	var req reqAsset.HostCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	now := time.Now()
	remark := req.Remark
	h := &model.AssetHost{
		GroupID:       req.GroupID,
		Name:          req.Name,
		IP:            req.IP,
		Port:          req.Port,
		OsType:        req.OsType,
		CloudProvider: req.CloudProvider,
		Username:      req.Username,
		AuthType:      req.AuthType,
		Tags:          req.Tags,
		Status:        "unknown",
		Remark:        &remark,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
	if req.Password != "" {
		h.Password = &req.Password
	}
	if req.PrivateKey != "" {
		h.PrivateKey = &req.PrivateKey
	}
	if err := c.hostMapper.Create(h); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": h.ID}})
}

// UpdateHost 编辑主机
func (c *HostController) UpdateHost(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqAsset.HostUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	h, err := c.hostMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "主机不存在")
		return
	}
	now := time.Now()
	h.GroupID = req.GroupID
	h.Name = req.Name
	h.IP = req.IP
	h.Port = req.Port
	h.OsType = req.OsType
	h.CloudProvider = req.CloudProvider
	h.Username = req.Username
	h.AuthType = req.AuthType
	h.Tags = req.Tags
	h.Remark = &req.Remark
	h.UpdatedAt = &now
	if req.Password != "" {
		h.Password = &req.Password
	}
	if req.PrivateKey != "" {
		h.PrivateKey = &req.PrivateKey
	}
	if err := c.hostMapper.Update(h); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// DeleteHost 删除主机
func (c *HostController) DeleteHost(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.hostMapper.SoftDelete(id); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// BatchDeleteHosts 批量删除主机
func (c *HostController) BatchDeleteHosts(ctx *gin.Context) {
	var req reqAsset.BatchDeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	if err := c.hostMapper.BatchDelete(req.IDs); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// GetHostStats 主机统计
func (c *HostController) GetHostStats(ctx *gin.Context) {
	groupIDStr := ctx.Query("group_id")
	var groupID uint64
	if groupIDStr != "" {
		groupID, _ = strconv.ParseUint(groupIDStr, 10, 64)
	}
	stats, err := c.hostMapper.Stats(groupID)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": stats})
}

func parseUint64Param(ctx *gin.Context, key string) (uint64, error) {
	return strconv.ParseUint(ctx.Param(key), 10, 64)
}
