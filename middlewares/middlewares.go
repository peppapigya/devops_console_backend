// 中间件层
package middlewares

import (
	"strconv"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(r *gin.Context) {
		// 这里可以添加身份验证逻辑
		r.Next()
	}
}

// InstanceAuth 实例认证中间件
func InstanceAuth() gin.HandlerFunc {
	return func(r *gin.Context) {
		// 从请求中获取实例ID，这里简化处理，实际应该从token或其他认证信息中获取
		// 首先尝试从URL参数中获取
		instanceIDStr := r.Query("instance_id")
		if instanceIDStr == "" {
			// 如果URL参数中没有，尝试从header中获取
			instanceIDStr = r.GetHeader("X-Instance-ID")
		}
		
		var instanceID uint
		if instanceIDStr != "" {
			// 如果找到了实例ID，转换为uint
			if id, err := strconv.ParseUint(instanceIDStr, 10, 32); err == nil {
				instanceID = uint(id)
			} else {
				// 转换失败，使用默认值
				instanceID = 1
			}
		} else {
			// 如果没有找到实例ID，使用默认值
			instanceID = 1
		}
		
		r.Set("instance_id", instanceID)
		r.Next()
	}
}
