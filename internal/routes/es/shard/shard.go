package shard

import (
	shardController "devops-console-backend/internal/controllers/es/shard"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册集群相关路由

func RegisterRoutes(shardGroup *gin.RouterGroup) {

	// 获取分片信息

	shardGroup.GET("/info", shardController.GetShardInfo)

	// 获取分片统计信息

	shardGroup.GET("/stats", shardController.GetShardStats)

	// 手动分配分片

	shardGroup.POST("/allocate", shardController.AllocateShard)

	// 移动分片

	shardGroup.POST("/move", shardController.MoveShard)

	// 迁移分片到另一个实例

	shardGroup.POST("/migrate-to-instance", shardController.MigrateShardToAnotherInstance)

	// 集群重路由操作

	shardGroup.POST("/cluster-reroute", shardController.ClusterReroute)

}

func RegisterSubRouter(g *gin.RouterGroup) {
	shardGroup := g.Group("/shard")
	RegisterRoutes(shardGroup)
}
