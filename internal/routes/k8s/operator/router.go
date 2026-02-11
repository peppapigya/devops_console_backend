package operator

import (
	"devops-console-backend/internal/controllers/k8s/operator"

	"github.com/gin-gonic/gin"
)

type OperatorRoute struct{}

func NewOperatorRoute() *OperatorRoute {
	return &OperatorRoute{}
}

func (r *OperatorRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	oc := operator.NewOperatorController()
	opGroup := apiGroup.Group("/k8s/operator")
	{
		opGroup.GET("/subscription/list/:namespace", oc.GetSubscriptionList)
		opGroup.GET("/subscription/detail/:namespace/:name", oc.GetSubscriptionDetail)
		opGroup.DELETE("/subscription/delete/:namespace/:name", oc.DeleteSubscription)
	}
}
