package instance

import (
	"devops-console-backend/models"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/logs"
	"devops-console-backend/repositories"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Get 获取实例详情
// @Summary 获取Elasticsearch实例详情
// @Description 根据ID获取指定Elasticsearch实例的详细信息
// @Tags instance
// @Accept json
// @Produce json
// @Param id query int true "实例ID"
// @Success 200 {object} common.ReturnData "成功"
// @Failure 400 {object} common.ReturnData "请求参数错误"
// @Failure 404 {object} common.ReturnData "实例不存在"
// @Failure 500 {object} common.ReturnData "服务器内部错误"
// @Router /api/instance/get [get]
func Get(r *gin.Context) {
	helper := utils.NewResponseHelper(r)

	// 获取并验证ID参数
	idStr := r.Query("id")
	if idStr == "" {
		helper.LogAndBadRequest("获取实例详情：缺少实例ID参数", nil)
		return
	}

	id64, err := strconv.ParseInt(idStr, 10, 64)
	id := uint(id64)
	if err != nil {
		helper.LogAndBadRequest("获取实例详情：无效的实例ID格式", map[string]interface{}{
			"id":    idStr,
			"error": err.Error(),
		})
		return
	}

	// 查询实例详情
	instanceDetail, err := getInstanceDetail(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.NotFound("实例不存在")
		} else {
			helper.DatabaseError("查询实例详情失败: " + err.Error())
		}
		return
	}

	// 不再添加地址前缀，保持原始地址格式
	// 地址格式将在前端根据https_enabled状态动态添加协议前缀

	logs.Info(map[string]interface{}{"id": id, "name": instanceDetail.ResourceName}, "获取实例详情成功")
	helper.SuccessWithData("获取成功", "instance", instanceDetail)
}

// getInstanceDetail 获取实例详情
func getInstanceDetail(id uint) (*models.ResourceDetail, error) {
	instanceDetailRepo := repositories.NewInstanceDetailRepository()
	instanceDetail, err := instanceDetailRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return instanceDetail, nil
}
