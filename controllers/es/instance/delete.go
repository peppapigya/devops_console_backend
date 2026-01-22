package instance

import (
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/logs"
	"devops-console-backend/repositories"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Delete 删除Elasticsearch实例
// @Summary 删除Elasticsearch实例
// @Description 根据ID删除指定的Elasticsearch实例
// @Tags instance
// @Accept json
// @Produce json
// @Param id query int true "实例ID"
// @Success 200 {object} common.ReturnData "成功"
// @Failure 400 {object} common.ReturnData "请求参数错误"
// @Failure 404 {object} common.ReturnData "实例不存在"
// @Failure 500 {object} common.ReturnData "服务器内部错误"
// @Router /api/instance/delete [get]
func Delete(r *gin.Context) {
	helper := utils.NewResponseHelper(r)

	// 获取并验证ID参数
	idStr := r.Query("id")
	if idStr == "" {
		helper.LogAndBadRequest("删除实例：缺少实例ID参数", nil)
		return
	}

	id64, err := strconv.ParseInt(idStr, 10, 64)
	id := uint(id64)
	if err != nil {
		helper.LogAndBadRequest("删除实例：无效的实例ID格式", map[string]interface{}{
			"id":    idStr,
			"error": err.Error(),
		})
		return
	}

	// 使用GORM删除实例
	instanceRepo := repositories.NewInstanceRepository()

	// 先获取实例信息用于日志记录
	instance, err := instanceRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.NotFound("实例不存在")
		} else {
			helper.DatabaseError("检查实例存在性失败: " + err.Error())
		}
		return
	}

	// 执行删除操作（GORM内部会处理事务）
	if err := instanceRepo.Delete(id); err != nil {
		helper.TransactionError("删除实例", err.Error())
		return
	}

	// 记录成功日志并返回响应
	logs.Info(map[string]interface{}{
		"id":            id,
		"instance_name": instance.Name,
	}, "删除实例成功")

	helper.Success("删除成功", map[string]interface{}{
		"deleted_id":   id,
		"deleted_name": instance.Name,
	})
}
