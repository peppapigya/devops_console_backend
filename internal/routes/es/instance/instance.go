package instance

import (
	common2 "devops-console-backend/internal/controllers/common"
	instance3 "devops-console-backend/internal/controllers/es/instance"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册集群相关路由
func RegisterRoutes(instanceGroup *gin.RouterGroup) {
	instanceGroup.POST("/add", instance3.Add)
	instanceGroup.POST("/update", instance3.Update)
	instanceGroup.GET("/delete", instance3.Delete)
	instanceGroup.GET("/get", instance3.Get)
	instanceGroup.GET("/list", instance3.List)
	instanceGroup.GET("/instance-types", common2.GetInstanceTypes)
	instanceGroup.POST("/test-connection", common2.TestConnection)
	instanceGroup.GET("/test-history", common2.GetTestHistory)
	instanceGroup.GET("/today-test-stats", common2.GetTodayTestStats)
}

func RegisterSubRouter(g *gin.RouterGroup) {
	instanceGroup := g.Group("/instance")
	RegisterRoutes(instanceGroup)
}
