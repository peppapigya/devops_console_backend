package backup

import (
	"devops-console-backend/internal/controllers/es/backup"

	"github.com/gin-gonic/gin"
)

// RegisterBackupRoutes 注册集群相关路由
func RegisterBackupRoutes(backupGroup *gin.RouterGroup) {
	// 仓库管理
	backupGroup.POST("/repository", backup.CreateRepository)
	backupGroup.DELETE("/repository", backup.DeleteRepository)
	backupGroup.GET("/repositories", backup.ListRepositories)

	// 快照管理
	backupGroup.POST("/snapshot", backup.CreateSnapshot)
	backupGroup.DELETE("/snapshot", backup.DeleteSnapshot)
	backupGroup.GET("/snapshots", backup.ListSnapshots)
	backupGroup.GET("/snapshot/status", backup.GetSnapshotStatus)

	// 恢复管理
	backupGroup.POST("/restore", backup.RestoreSnapshot)
	backupGroup.GET("/restore/status", backup.GetRestoreStatus)
}

func RegisterSubRouter(g *gin.RouterGroup) {
	esGroup := g.Group("/backup")
	RegisterBackupRoutes(esGroup)
}
