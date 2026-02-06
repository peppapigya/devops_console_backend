package network

import (
	"devops-console-backend/internal/controllers/k8s/network"

	"github.com/gin-gonic/gin"
)

type NetworkRoute struct {
	ingressController *network.IngressController
}

func NewNetworkRoute() *NetworkRoute {
	return &NetworkRoute{
		ingressController: network.NewIngressController(),
	}
}

func (r *NetworkRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	// Ingress
	ingressGroup := apiGroup.Group("/k8s/ingress")
	{
		ingressGroup.GET("/list/:namespace", r.ingressController.GetIngressList)
		ingressGroup.GET("/list/all", r.ingressController.GetIngressList)
		ingressGroup.GET("/detail/:namespace/:name", r.ingressController.GetIngressDetail)
		ingressGroup.POST("/create", r.ingressController.CreateIngress)
		ingressGroup.PUT("/update/:namespace/:name", r.ingressController.UpdateIngress)
		ingressGroup.DELETE("/delete/:namespace/:name", r.ingressController.DeleteIngress)
	}
}
