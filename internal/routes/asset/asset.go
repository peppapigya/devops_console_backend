package asset

import (
	assetCtrl "devops-console-backend/internal/controllers/asset"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/pkg/configs"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterAssetRouters 注册资产管理路由
func RegisterAssetRouters(apiGroup *gin.RouterGroup, externalDB ...*gorm.DB) {
	var db *gorm.DB
	if len(externalDB) > 0 && externalDB[0] != nil {
		db = externalDB[0]
	} else {
		db = configs.NewDB()
	}

	groupMapper := mapper.NewAssetHostGroupMapper(db)
	hostMapper := mapper.NewAssetHostMapper(db)
	hostCtrl := assetCtrl.NewHostController(groupMapper, hostMapper)

	assetGroup := apiGroup.Group("/asset")
	{
		// ===== 主机分组 =====
		groupRoute := assetGroup.Group("/host-groups")
		{
			groupRoute.GET("", hostCtrl.ListGroups)
			groupRoute.POST("", hostCtrl.CreateGroup)
			groupRoute.PUT("/:id", hostCtrl.UpdateGroup)
			groupRoute.DELETE("/:id", hostCtrl.DeleteGroup)
		}

		// ===== 主机管理 =====
		hostRoute := assetGroup.Group("/hosts")
		{
			hostRoute.GET("", hostCtrl.ListHosts)
			hostRoute.GET("/stats", hostCtrl.GetHostStats)
			hostRoute.POST("", hostCtrl.CreateHost)
			hostRoute.PUT("/:id", hostCtrl.UpdateHost)
			hostRoute.DELETE("/:id", hostCtrl.DeleteHost)
			hostRoute.DELETE("/batch", hostCtrl.BatchDeleteHosts)
		}
	}
}
