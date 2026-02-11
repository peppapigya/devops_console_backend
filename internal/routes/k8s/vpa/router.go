package vpa

import (
	"devops-console-backend/internal/controllers/k8s/vpa"

	"github.com/gin-gonic/gin"
)

type VpaRoute struct{}

func NewVpaRoute() *VpaRoute {
	return &VpaRoute{}
}

func (r *VpaRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	vc := vpa.NewVPAController()
	vpaGroup := apiGroup.Group("/k8s/vpa")
	{
		vpaGroup.GET("/list/:namespace", vc.GetVPAList)
		vpaGroup.GET("/detail/:namespace/:name", vc.GetVPADetail)
		vpaGroup.DELETE("/delete/:namespace/:name", vc.DeleteVPA)
	}
}
