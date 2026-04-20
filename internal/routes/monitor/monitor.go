package monitor

import (
	"devops-console-backend/internal/agent"
	alertingCtrl "devops-console-backend/internal/controllers/alerting"
	monitorCtrl "devops-console-backend/internal/controllers/monitor"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/middlewares"
	"devops-console-backend/internal/services/alerting"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterMonitorRouters(router *gin.RouterGroup, db *gorm.DB) {
	prometheusController := monitorCtrl.NewPrometheusController(db)
	alertmanagerController := alertingCtrl.NewAlertmanagerController(alerting.NewOrchestratorFromDB(db, agent.RunAgentLoop))

	router.Any("/monitor/prometheus/:instanceId/*path", middlewares.Authenticate(), prometheusController.Proxy)
	router.POST("/monitor/alertmanager/webhook", alertmanagerController.Webhook)

	customGroup := router.Group("/monitor/custom")
	customGroup.Use(middlewares.Authenticate())
	customGroup.GET("", monitorCtrl.ListCustomMonitors)
	customGroup.POST("", monitorCtrl.CreateCustomMonitor)
	customGroup.PUT("/:id", monitorCtrl.UpdateCustomMonitor)
	customGroup.DELETE("/:id", monitorCtrl.DeleteCustomMonitor)

	domainCtrl := monitorCtrl.NewDomainController(
		mapper.NewDomainMapper(db),
		mapper.NewSslCertMapper(db),
		mapper.NewDnsProviderMapper(db),
	)
	domainGroup := router.Group("/monitor/domain")
	domainGroup.Use(middlewares.Authenticate())
	{
		domainGroup.GET("/stats", domainCtrl.Stats)
		domainGroup.GET("", domainCtrl.ListDomains)
		domainGroup.POST("", domainCtrl.CreateDomain)
		domainGroup.PUT("/:id", domainCtrl.UpdateDomain)
		domainGroup.DELETE("/:id", domainCtrl.DeleteDomain)
		domainGroup.PUT("/:id/toggle", domainCtrl.ToggleDomain)
		domainGroup.GET("/ssl-certs", domainCtrl.ListSslCerts)
		domainGroup.POST("/ssl-certs/apply", domainCtrl.ApplySslCert)
		domainGroup.POST("/ssl-certs/upload", domainCtrl.UploadSslCert)
		domainGroup.DELETE("/ssl-certs/:id", domainCtrl.DeleteSslCert)
		domainGroup.GET("/ssl-certs/:id/download", domainCtrl.DownloadSslCert)
		domainGroup.GET("/dns-providers", domainCtrl.ListDnsProviders)
		domainGroup.POST("/dns-providers", domainCtrl.CreateDnsProvider)
		domainGroup.PUT("/dns-providers/:id", domainCtrl.UpdateDnsProvider)
		domainGroup.DELETE("/dns-providers/:id", domainCtrl.DeleteDnsProvider)
		domainGroup.POST("/dns-providers/:id/test", domainCtrl.TestDnsProvider)
	}

	incidentCtrl := monitorCtrl.NewIncidentController(mapper.NewIncidentMapper(db))
	incidentGroup := router.Group("/monitor/incident")
	incidentGroup.Use(middlewares.Authenticate())
	{
		incidentGroup.GET("/stats", incidentCtrl.Stats)
		incidentGroup.GET("/business-lines", incidentCtrl.GetBusinessLines)
		incidentGroup.GET("", incidentCtrl.List)
		incidentGroup.GET("/:id", incidentCtrl.GetByID)
		incidentGroup.POST("", incidentCtrl.Create)
		incidentGroup.PUT("/:id", incidentCtrl.Update)
		incidentGroup.DELETE("/:id", incidentCtrl.Delete)
		incidentGroup.PUT("/:id/status", incidentCtrl.UpdateStatus)
	}
}
