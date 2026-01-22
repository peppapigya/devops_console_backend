package instance

import (
	"devops-console-backend/controllers/common"
	instance2 "devops-console-backend/controllers/es/instance"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册集群相关路由
func RegisterRoutes(instanceGroup *gin.RouterGroup) {
	instanceGroup.POST("/add", instance2.Add)
	instanceGroup.POST("/update", instance2.Update)
	instanceGroup.GET("/delete", instance2.Delete)
	instanceGroup.GET("/get", instance2.Get)
	instanceGroup.GET("/list", instance2.List)
	instanceGroup.GET("/instance-types", common.GetInstanceTypes)
	instanceGroup.POST("/test-connection", common.TestConnection)
	instanceGroup.GET("/test-history", common.GetTestHistory)
	instanceGroup.GET("/today-test-stats", common.GetTodayTestStats)
}

func RegisterSubRouter(g *gin.RouterGroup) {
	instanceGroup := g.Group("/instance")
	RegisterRoutes(instanceGroup)
}
