package instance

import (
	"devops-console-backend/database"
	"devops-console-backend/models"
	"devops-console-backend/models/request"
	"devops-console-backend/models/response"
	"devops-console-backend/repositories"
	"devops-console-backend/utils"
	"devops-console-backend/utils/logs"

	"github.com/gin-gonic/gin"
)

// List 获取实例列表
// @Summary 获取Elasticsearch实例列表
// @Description 获取Elasticsearch实例列表，支持分页和过滤
// @Tags instance
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "实例状态过滤"
// @Param type_name query string false "实例类型过滤"
// @Param keyword query string false "关键词搜索"
// @Success 200 {object} common.ReturnData "成功"
// @Failure 400 {object} common.ReturnData "请求参数错误"
// @Failure 500 {object} common.ReturnData "服务器内部错误"
// @Router /api/instance/list [get]
func List(r *gin.Context) {
	helper := utils.NewResponseHelper(r)

	// 解析查询参数
	var req request.InstanceListRequest
	if err := r.ShouldBindQuery(&req); err != nil {
		helper.BadRequest("参数解析失败: " + err.Error())
		return
	}

	// 设置默认值和验证
	req = setDefaultsAndValidate(&req)

	// 构建过滤条件
	filters := make(map[string]interface{})
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.TypeName != "" {
		filters["type_name"] = req.TypeName
	}
	if req.Keyword != "" {
		filters["name"] = req.Keyword
	}

	// 使用GORM查询
	instanceRepo := repositories.NewInstanceRepository()
	offset := (req.Page - 1) * req.PageSize

	instances, total, err := instanceRepo.GetWithPagination(offset, req.PageSize, filters)
	if err != nil {
		helper.DatabaseError("查询集群列表失败: " + err.Error())
		return
	}

	// 构建实例类型映射 - 手动查询实例类型
	instanceTypes := make(map[uint64]string)
	var instanceTypeIDs []uint
	for _, instance := range instances {
		instanceTypeIDs = append(instanceTypeIDs, instance.InstanceTypeID)
	}

	var instanceTypesList []models.InstanceType
	if err := database.GORMDB.Where("id IN ?", instanceTypeIDs).Find(&instanceTypesList).Error; err == nil {
		for _, instanceType := range instanceTypesList {
			instanceTypes[uint64(instanceType.ID)] = instanceType.TypeName
		}
	}

	logs.Debug(nil, "查询集群列表成功")
	helper.Success("查询成功", map[string]interface{}{
		"list": response.NewInstanceListResponse(req.Page, req.PageSize, total, instances, instanceTypes),
	})
}

// setDefaultsAndValidate 设置默认值和验证参数
func setDefaultsAndValidate(req *request.InstanceListRequest) request.InstanceListRequest {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	return *req
}
