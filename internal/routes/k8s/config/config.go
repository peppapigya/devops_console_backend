package config

import (
	"devops-console-backend/internal/controllers/k8s/config"

	"github.com/gin-gonic/gin"
)

type ConfigRoute struct {
	configMapController *config.ConfigMapController
	secretController    *config.SecretController
}

func NewConfigRoute() *ConfigRoute {
	return &ConfigRoute{
		configMapController: config.NewConfigMapController(),
		secretController:    config.NewSecretController(),
	}
}

func (r *ConfigRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	// ConfigMap
	configMapGroup := apiGroup.Group("/k8s/configmap")
	{
		configMapGroup.GET("/list/:namespace", r.configMapController.GetConfigMapList)
		configMapGroup.GET("/list/all", r.configMapController.GetConfigMapList)
		configMapGroup.GET("/detail/:namespace/:name", r.configMapController.GetConfigMapDetail)
		configMapGroup.POST("/create", r.configMapController.CreateConfigMap)
		configMapGroup.PUT("/update/:namespace/:name", r.configMapController.UpdateConfigMap)
		configMapGroup.DELETE("/delete/:namespace/:name", r.configMapController.DeleteConfigMap)
	}

	// Secret
	secretGroup := apiGroup.Group("/k8s/secret")
	{
		secretGroup.GET("/list/:namespace", r.secretController.GetSecretList)
		secretGroup.GET("/list/all", r.secretController.GetSecretList)
		secretGroup.GET("/detail/:namespace/:name", r.secretController.GetSecretDetail)
		secretGroup.POST("/create", r.secretController.CreateSecret)
		secretGroup.PUT("/update/:namespace/:name", r.secretController.UpdateSecret)
		secretGroup.DELETE("/delete/:namespace/:name", r.secretController.DeleteSecret)
	}
}
