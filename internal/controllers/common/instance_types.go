package common

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils/logs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetInstanceTypes 获取所有实例类型
// @Summary 获取实例类型列表
// @Description 获取所有可用的实例类型
// @Tags instance
// @Accept json
// @Produce json
// @Success 200 {object} common.ReturnData "成功"
// @Failure 500 {object} common.ReturnData "服务器内部错误"
// @Router /api/instance/instance-types [get]
func GetInstanceTypes(c *gin.Context) {
	logs.Info(map[string]interface{}{}, "获取实例类型列表")

	// 使用GORM查询所有实例类型
	instanceTypeRepo := configs.NewInstanceTypeRepository()
	instanceTypes, err := instanceTypeRepo.GetAll()
	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "查询实例类型失败")
		c.JSON(http.StatusInternalServerError, common.ReturnData{
			Status:  http.StatusInternalServerError,
			Message: "查询实例类型失败",
			Data:    nil,
		})
		return
	}

	logs.Info(map[string]interface{}{"count": len(instanceTypes)}, "获取实例类型成功")

	c.JSON(http.StatusOK, common.ReturnData{
		Status:  http.StatusOK,
		Message: "获取实例类型成功",
		Data:    map[string]interface{}{"instance_types": instanceTypes},
	})
}
