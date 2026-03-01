package monitor

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	reqIncident "devops-console-backend/internal/dal/request/incident"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// IncidentController 故障管理控制器
type IncidentController struct {
	incidentMapper *mapper.IncidentMapper
}

func NewIncidentController(im *mapper.IncidentMapper) *IncidentController {
	return &IncidentController{incidentMapper: im}
}

// Stats 统计故障数据
func (c *IncidentController) Stats(ctx *gin.Context) {
	stats, err := c.incidentMapper.Stats()
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": stats})
}

// List 分页查询故障列表
func (c *IncidentController) List(ctx *gin.Context) {
	var req reqIncident.IncidentPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	total, list, err := c.incidentMapper.ListPage(req.Page, req.PageSize, req.BusinessLine, req.Level, req.Status, req.Dept)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

// GetByID 查询单条故障
func (c *IncidentController) GetByID(ctx *gin.Context) {
	id, err := parseIncidentUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	inc, err := c.incidentMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "故障记录不存在")
		return
	}
	common.Success(ctx, gin.H{"data": inc})
}

// Create 新建故障记录
func (c *IncidentController) Create(ctx *gin.Context) {
	var req reqIncident.IncidentCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	alertTime, err := time.ParseInLocation("2006-01-02 15:04:05", req.AlertTime, time.Local)
	if err != nil {
		common.FailWithMsg(ctx, "告警时间格式错误，请使用 YYYY-MM-DD HH:mm:ss")
		return
	}
	now := time.Now()
	status := req.Status
	if status == "" {
		status = "pending"
	}
	freq := req.Frequency
	if freq == "" {
		freq = "偶发"
	}
	detail := req.Detail
	inc := &model.MonitorIncident{
		AlertTime:    alertTime,
		BusinessLine: req.BusinessLine,
		Level:        req.Level,
		Frequency:    freq,
		AlertDesc:    req.AlertDesc,
		Detail:       &detail,
		Dept:         req.Dept,
		Handler:      req.Handler,
		Status:       status,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}
	if req.ResolvedAt != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", req.ResolvedAt, time.Local); err == nil {
			inc.ResolvedAt = &t
		}
	}
	if err := c.incidentMapper.Create(inc); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": inc.ID}})
}

// Update 编辑故障记录
func (c *IncidentController) Update(ctx *gin.Context) {
	id, err := parseIncidentUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqIncident.IncidentUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	inc, err := c.incidentMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "故障记录不存在")
		return
	}
	now := time.Now()
	inc.BusinessLine = req.BusinessLine
	inc.Level = req.Level
	if req.Frequency != "" {
		inc.Frequency = req.Frequency
	}
	inc.AlertDesc = req.AlertDesc
	detail := req.Detail
	inc.Detail = &detail
	inc.Dept = req.Dept
	inc.Handler = req.Handler
	if req.Status != "" {
		inc.Status = req.Status
	}
	inc.UpdatedAt = &now
	if req.AlertTime != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", req.AlertTime, time.Local); err == nil {
			inc.AlertTime = t
		}
	}
	if req.ResolvedAt != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", req.ResolvedAt, time.Local); err == nil {
			inc.ResolvedAt = &t
		}
	}
	if err := c.incidentMapper.Update(inc); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// Delete 删除故障记录
func (c *IncidentController) Delete(ctx *gin.Context) {
	id, err := parseIncidentUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.incidentMapper.SoftDelete(id); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// UpdateStatus 更新处理状态
func (c *IncidentController) UpdateStatus(ctx *gin.Context) {
	id, err := parseIncidentUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqIncident.IncidentStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	fields := map[string]interface{}{"status": req.Status}
	if req.Status == "done" {
		now := time.Now()
		fields["resolved_at"] = now
	}
	if err := c.incidentMapper.UpdateFields(id, fields); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// GetBusinessLines 获取业务线列表
func (c *IncidentController) GetBusinessLines(ctx *gin.Context) {
	lines, err := c.incidentMapper.ListBusinessLines()
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": lines})
}

func parseIncidentUint64Param(ctx *gin.Context, key string) (uint64, error) {
	return strconv.ParseUint(ctx.Param(key), 10, 64)
}
