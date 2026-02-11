package crd

import (
	"devops-console-backend/internal/controllers/k8s/crd"

	"github.com/gin-gonic/gin"
)

type CrdRoute struct{}

func NewCrdRoute() *CrdRoute {
	return &CrdRoute{}
}

func (r *CrdRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	cc := crd.NewCRDController()
	crdGroup := apiGroup.Group("/k8s/crd")
	{
		crdGroup.GET("/list", cc.GetCRDList)
		crdGroup.GET("/detail/:name", cc.GetCRDDetail)
		crdGroup.DELETE("/delete/:name", cc.DeleteCRD)
	}
}
