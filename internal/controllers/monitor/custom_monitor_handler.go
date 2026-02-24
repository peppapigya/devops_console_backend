package monitor

import (
	"devops-console-backend/internal/models/request"
	"devops-console-backend/internal/services"
	"devops-console-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListCustomMonitors(c *gin.Context) {
	accountID := utils.GetUserIdFromContext(c)
	helper := utils.NewResponseHelper(c)
	if accountID == 0 {
		helper.Error(http.StatusUnauthorized, "未授权")
		return
	}

	targetType := c.Query("target_type")
	monitors, err := services.ListCustomMonitors(accountID, targetType)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}

	helper.SuccessWithData("获取成功", "list", monitors)
}

func CreateCustomMonitor(c *gin.Context) {
	accountID := utils.GetUserIdFromContext(c)
	helper := utils.NewResponseHelper(c)
	if accountID == 0 {
		helper.Error(http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req request.CreateCustomMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(err.Error())
		return
	}

	monitor, err := services.CreateCustomMonitor(accountID, req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}

	helper.SuccessWithData("创建成功", "monitor", monitor)
}

func UpdateCustomMonitor(c *gin.Context) {
	accountID := utils.GetUserIdFromContext(c)
	helper := utils.NewResponseHelper(c)
	if accountID == 0 {
		helper.Error(http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		helper.BadRequest("invalid id format")
		return
	}

	var req request.UpdateCustomMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(err.Error())
		return
	}

	monitor, err := services.UpdateCustomMonitor(uint(accountID), uint(id), req)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}

	helper.SuccessWithData("更新成功", "monitor", monitor)
}

func DeleteCustomMonitor(c *gin.Context) {
	accountID := utils.GetUserIdFromContext(c)
	helper := utils.NewResponseHelper(c)
	if accountID == 0 {
		helper.Error(http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		helper.BadRequest("invalid id format")
		return
	}

	if err := services.DeleteCustomMonitor(uint(accountID), uint(id)); err != nil {
		helper.InternalError(err.Error())
		return
	}

	helper.Success("删除成功")
}
