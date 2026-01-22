package elasticsearch

import (
	"devops-console-backend/controllers/es/dev_console"
	"devops-console-backend/controllers/es/elasticsearch"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册Elasticsearch相关路由
func RegisterRoutes(esGroup *gin.RouterGroup) {
	// 集群状态相关路由
	esGroup.GET("/cluster/status", elasticsearch.GetClusterStatus)
	esGroup.GET("/cluster/health", elasticsearch.GetClusterHealthHandler)
	esGroup.GET("/cluster/info", elasticsearch.GetClusterInfoHandler)
	esGroup.GET("/cluster/full-info", elasticsearch.GetFullClusterInfo)

	// 开发者控制台路由
	esGroup.Any("/dev/*path", dev_console.DevConsoleHandler)                   // 支持所有HTTP方法的原生代理方式
	esGroup.POST("/dev-console/execute", dev_console.ExecuteDevConsoleRequest) // JSON请求体方式
	esGroup.GET("/dev-console/validate", dev_console.ValidateInstance)         // 验证实例连接
	esGroup.GET("/dev-console/info", dev_console.GetInstanceInfo)              // 获取实例详细信息
}

// RegisterSubRouter 注册Elasticsearch子路由组
func RegisterSubRouter(g *gin.RouterGroup) {
	esGroup := g.Group("/elasticsearch")
	RegisterRoutes(esGroup)
}
