package utils

import (
	"devops-console-backend/common"
	"devops-console-backend/pkg/utils/logs"

	"github.com/gin-gonic/gin"
)

// ResponseHelper 统一响应处理助手
type ResponseHelper struct {
	GinContext *gin.Context
}

// NewResponseHelper 创建响应助手
func NewResponseHelper(c *gin.Context) *ResponseHelper {
	return &ResponseHelper{
		GinContext: c,
	}
}

// Success 成功响应
func (rh *ResponseHelper) Success(message string, data ...map[string]interface{}) {
	response := common.NewReturnData()
	response.Message = message
	if len(data) > 0 {
		for _, d := range data {
			for k, v := range d {
				response.Data[k] = v
			}
		}
	}
	rh.GinContext.JSON(200, response)
}

// SuccessWithData 成功响应（带数据）
func (rh *ResponseHelper) SuccessWithData(message string, key string, value interface{}) {
	response := common.NewReturnData()
	response.Message = message
	response.Data[key] = value
	rh.GinContext.JSON(200, response)
}

// Error 错误响应
func (rh *ResponseHelper) Error(status int, message string) {
	response := common.NewReturnData()
	response.Status = status
	response.Message = message
	rh.GinContext.JSON(200, response)
}

// BadRequest 400错误响应
func (rh *ResponseHelper) BadRequest(message string) {
	rh.Error(400, message)
}

// NotFound 404错误响应
func (rh *ResponseHelper) NotFound(message string) {
	rh.Error(404, message)
}

// DatabaseError 数据库错误响应
func (rh *ResponseHelper) DatabaseError(message string) {
	rh.Error(500, "数据库操作失败: "+message)
}

// LogAndBadRequest 记录警告日志并返回400错误
func (rh *ResponseHelper) LogAndBadRequest(message string, logData map[string]interface{}) {
	logs.Warning(logData, message)
	rh.BadRequest(message)
}

// TransactionError 事务错误响应
func (rh *ResponseHelper) TransactionError(operation, step string) {
	logs.Error(map[string]interface{}{
		"operation": operation,
		"step":      step,
	}, operation+"失败: "+step)
	rh.Error(500, operation+"失败")
}

// InternalError 500错误响应（兼容性方法）
func (rh *ResponseHelper) InternalError(message string) {
	rh.Error(500, message)
}

// ValidationError 参数验证错误响应（兼容性方法）
func (rh *ResponseHelper) ValidationError(message string) {
	rh.BadRequest(message)
}
