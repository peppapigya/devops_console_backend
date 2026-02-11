package replicationcontroller

import (
	"devops-console-backend/internal/controllers/k8s/replicationcontroller"

	"github.com/gin-gonic/gin"
)

type ReplicationControllerRoute struct{}

func NewReplicationControllerRoute() *ReplicationControllerRoute {
	return &ReplicationControllerRoute{}
}

func (r *ReplicationControllerRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	rc := replicationcontroller.NewReplicationControllerController()
	rcGroup := apiGroup.Group("/k8s/rc")
	{
		rcGroup.GET("/list/:namespace", rc.GetRCList)
		rcGroup.GET("/detail/:namespace/:name", rc.GetRCDetail)
		rcGroup.DELETE("/delete/:namespace/:name", rc.DeleteRC)
	}
}
