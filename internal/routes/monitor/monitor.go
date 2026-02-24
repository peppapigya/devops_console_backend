package monitor

import (
	"devops-console-backend/internal/controllers/monitor"
	"devops-console-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterMonitorRouters 注册监控相关的路由
func RegisterMonitorRouters(router *gin.RouterGroup, db *gorm.DB) {
	prometheusController := monitor.NewPrometheusController(db)

	// prometheus 代理路由: 允许任意后面带有路径的请求，只要带上 :instanceId
	// 我们加了 auth 中间件，保证了安全性
	// Any 会匹配 GET/POST 等所有方法，*path 会匹配后续所有的路径
	router.Any("/monitor/prometheus/:instanceId/*path", middlewares.Authenticate(), prometheusController.Proxy)

	// 自定义监控大盘路由
	customGroup := router.Group("/monitor/custom")
	customGroup.Use(middlewares.Authenticate()) // 添加认证中间件保护

	customGroup.GET("", monitor.ListCustomMonitors)
	customGroup.POST("", monitor.CreateCustomMonitor)
	customGroup.PUT("/:id", monitor.UpdateCustomMonitor)
	customGroup.DELETE("/:id", monitor.DeleteCustomMonitor)
}
