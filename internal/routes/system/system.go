package system

import (
	"github.com/gin-gonic/gin"
)

func RegisterSystemRouters(router *gin.RouterGroup) {
	RegisterLoginRoutes(router)
}
