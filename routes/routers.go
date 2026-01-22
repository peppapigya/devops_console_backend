// 路由层 管理程序的路由信息
package routers

import (
	"devops-console-backend/routes/es/backup"
	"devops-console-backend/routes/es/elasticsearch"
	"devops-console-backend/routes/es/indices"
	"devops-console-backend/routes/es/instance"
	"devops-console-backend/routes/es/node"
	"devops-console-backend/routes/es/shard"
	"devops-console-backend/routes/k8s"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRouters 注册路由的方法
func RegisterRouters(r *gin.Engine, db *gorm.DB) {
	// 路由配置
	apiGroup := r.Group("/api")
	elasticsearch.RegisterSubRouter(apiGroup)
	backup.RegisterSubRouter(apiGroup)
	instance.RegisterSubRouter(apiGroup)
	node.RegisterSubRouter(apiGroup)
	shard.RegisterSubRouter(apiGroup)
	indices.RegisterSubRouter(apiGroup)
	
	// 注册K8s模块路由
	k8s.RegisterK8sRoutes(apiGroup, db)
}
