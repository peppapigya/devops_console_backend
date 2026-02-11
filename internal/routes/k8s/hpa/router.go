package hpa

import (
	"devops-console-backend/internal/controllers/k8s/hpa"

	"github.com/gin-gonic/gin"
)

type HpaRoute struct{}

func NewHpaRoute() *HpaRoute {
	return &HpaRoute{}
}

func (r *HpaRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	hc := hpa.NewHPAController()
	hpaGroup := apiGroup.Group("/k8s/hpa")
	{
		hpaGroup.GET("/list/:namespace", hc.GetHPAList)
		hpaGroup.GET("/detail/:namespace/:name", hc.GetHPADetail)
		hpaGroup.DELETE("/delete/:namespace/:name", hc.DeleteHPA)
	}
}
