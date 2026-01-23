package cluster

import (
	"devops-console-backend/internal/controllers/k8s/cluster"

	"github.com/gin-gonic/gin"
)

// SetupClusterRoutes 设置集群相关路由
func SetupClusterRoutes(router *gin.RouterGroup) {
	clusterController := cluster.NewClusterController()

	clusterGroup := router.Group("/cluster")
	{
		// 获取集群列表
		clusterGroup.GET("/list", clusterController.GetClusterList)

		// 获取集群基本信息
		clusterGroup.GET("/info", clusterController.GetClusterInfo)

		// 获取集群指标数据
		clusterGroup.GET("/metrics", clusterController.GetClusterMetrics)

		// 获取节点列表
		clusterGroup.GET("/nodes", clusterController.GetNodeList)
	}
}
