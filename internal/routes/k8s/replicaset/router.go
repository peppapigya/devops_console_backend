package replicaset

import (
	"devops-console-backend/internal/controllers/k8s/replicaset"

	"github.com/gin-gonic/gin"
)

type ReplicaSetRoute struct{}

func NewReplicaSetRoute() *ReplicaSetRoute {
	return &ReplicaSetRoute{}
}

func (r *ReplicaSetRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	rc := replicaset.NewReplicaSetController()
	rsGroup := apiGroup.Group("/k8s/replicaset")
	{
		rsGroup.GET("/list/:namespace", rc.GetReplicaSetList)
		rsGroup.GET("/detail/:namespace/:name", rc.GetReplicaSetDetail)
		rsGroup.POST("/create", rc.CreateReplicaSet)
		rsGroup.DELETE("/delete/:namespace/:name", rc.DeleteReplicaSet)
	}
}
