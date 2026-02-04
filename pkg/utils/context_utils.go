package utils

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/pkg/utils/jwt"

	"github.com/gin-gonic/gin"
)

// GetUserInfoFromContext 从上下文中获取用户信息
func GetUserInfoFromContext(c *gin.Context) *jwt.Claims {
	claims, _ := c.Get(common.UserInfoKey)
	if claims == nil {
		common.Fail(c, common.UserNotExist)
		return nil
	}
	return claims.(*jwt.Claims)
}

// GetUserIdFromContext 从上下文中获取用户ID
func GetUserIdFromContext(c *gin.Context) int64 {
	return GetUserInfoFromContext(c).GetUserId()
}

func GetUserNameFromContext(c *gin.Context) string {
	return GetUserInfoFromContext(c).GetUserName()
}
