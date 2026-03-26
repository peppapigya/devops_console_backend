// 路由层 管理程序的路由信息
package routers

import (
	"devops-console-backend/internal/routes/asset"
	"devops-console-backend/internal/routes/cicd"
	"devops-console-backend/internal/routes/es/backup"
	"devops-console-backend/internal/routes/es/elasticsearch"
	"devops-console-backend/internal/routes/es/indices"
	"devops-console-backend/internal/routes/es/instance"
	"devops-console-backend/internal/routes/es/node"
	"devops-console-backend/internal/routes/es/shard"
	"devops-console-backend/internal/routes/helm"
	"devops-console-backend/internal/routes/k8s"
	"devops-console-backend/internal/routes/monitor"
	"devops-console-backend/internal/routes/system"
	"devops-console-backend/internal/routes/task_scheduler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRouters 注册路由的方法
func RegisterRouters(r *gin.Engine, db *gorm.DB) {
	// 路由配置
	apiGroup := r.Group("/api/v1")
	{
		elasticsearch.RegisterSubRouter(apiGroup)
		backup.RegisterSubRouter(apiGroup)
		instance.RegisterSubRouter(apiGroup)
		node.RegisterSubRouter(apiGroup)
		shard.RegisterSubRouter(apiGroup)
		indices.RegisterSubRouter(apiGroup)
		// 注册K8s模块路由
		k8s.RegisterK8sRoutes(apiGroup, db)
		// 注册Helm模块路由
		helmRoute := helm.NewHelmRoute(db)
		helmRoute.RegisterSubRouter(apiGroup)
		system.RegisterSystemRouters(apiGroup)

		// 注册监控(Prometheus等)模块路由
		monitor.RegisterMonitorRouters(apiGroup, db)

		// 注册资产管理模块路由
		asset.RegisterAssetRouters(apiGroup, db)

		// CiCd 模块
		cicd.RegisterCiCdRouters(apiGroup)

		// 任务调度模块
		task_scheduler.RegisterTaskSchedulerRouters(apiGroup, db)
	}
}
