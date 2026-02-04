package indices

import (
	"devops-console-backend/internal/controllers/es/indices"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册集群相关路由
func RegisterRoutes(indicesGroup *gin.RouterGroup) {
	indicesGroup.POST("/indexcreate", indices.Indexcreate)
	indicesGroup.GET("/catindices", indices.Catindices)
	indicesGroup.POST("/deleteindices", indices.Deleteindices)
	indicesGroup.POST("/updateindices", indices.Updateindices)

}

func RegisterSubRouter(g *gin.RouterGroup) {
	indicesGroup := g.Group("/indices")
	RegisterRoutes(indicesGroup)
}
