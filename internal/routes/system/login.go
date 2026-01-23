package system

import (
	"devops-console-backend/cmd/generate/wireInfo"

	"github.com/gin-gonic/gin"
)

func RegisterLoginRoutes(router *gin.RouterGroup) {
	loginController := wireInfo.InitializeLoginController()
	systemGroup := router.Group("/system")
	{
		systemGroup.POST("/login", loginController.Login)
	}
}
