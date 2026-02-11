package k8s

import (
	"devops-console-backend/internal/routes/k8s/cluster"
	"devops-console-backend/internal/routes/k8s/config"
	"devops-console-backend/internal/routes/k8s/crd"
	"devops-console-backend/internal/routes/k8s/cronjob"
	"devops-console-backend/internal/routes/k8s/daemonset"
	"devops-console-backend/internal/routes/k8s/deployment"
	"devops-console-backend/internal/routes/k8s/event"
	"devops-console-backend/internal/routes/k8s/hpa"
	"devops-console-backend/internal/routes/k8s/job"
	"devops-console-backend/internal/routes/k8s/namespace"
	"devops-console-backend/internal/routes/k8s/network"
	"devops-console-backend/internal/routes/k8s/node"
	"devops-console-backend/internal/routes/k8s/operator"
	"devops-console-backend/internal/routes/k8s/pod"
	"devops-console-backend/internal/routes/k8s/replicaset"
	"devops-console-backend/internal/routes/k8s/replicationcontroller"
	"devops-console-backend/internal/routes/k8s/service"
	"devops-console-backend/internal/routes/k8s/storage"
	"devops-console-backend/internal/routes/k8s/vpa"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterK8sRoutes 注册K8s相关路由
func RegisterK8sRoutes(apiGroup *gin.RouterGroup, db *gorm.DB) {
	// 注册集群路由
	cluster.SetupClusterRoutes(apiGroup)

	// 注册Pod路由
	podRoute := pod.NewPodRoute()
	podRoute.RegisterSubRouter(apiGroup)

	// 注册Node路由
	nodeRoute := node.NewNodeRoute()
	nodeRoute.RegisterSubRouter(apiGroup)

	// 注册Deployment路由
	deploymentRoute := deployment.NewDeploymentRoute()
	deploymentRoute.RegisterSubRouter(apiGroup)

	// 注册CronJob路由
	cronjobRoute := cronjob.NewCronJobRoute()
	cronjobRoute.RegisterSubRouter(apiGroup)

	// 注册DaemonSet路由
	daemonsetRoute := daemonset.NewDaemonSetRoute()
	daemonsetRoute.RegisterSubRouter(apiGroup)

	// 注册Job路由
	jobRoute := job.NewJobRoute()
	jobRoute.RegisterSubRouter(apiGroup)

	// 注册Namespace路由
	namespaceRoute := namespace.NewNamespaceRoute()
	namespaceRoute.RegisterSubRouter(apiGroup)

	// 注册Service路由
	serviceRoute := service.NewServiceRoute()
	serviceRoute.RegisterSubRouter(apiGroup)

	// 注册Storage路由
	storageRoute := storage.NewStorageRoute()
	storageRoute.RegisterSubRouter(apiGroup)

	// 注册Network路由
	networkRoute := network.NewNetworkRoute()
	networkRoute.RegisterSubRouter(apiGroup)

	// 注册Config路由
	configRoute := config.NewConfigRoute()
	configRoute.RegisterSubRouter(apiGroup)

	// 注册Event路由
	eventRoute := event.NewEventRoute()
	eventRoute.RegisterSubRouter(apiGroup)

	// 注册ReplicaSet路由
	rsRoute := replicaset.NewReplicaSetRoute()
	rsRoute.RegisterSubRouter(apiGroup)

	// 注册ReplicationController路由
	rcRoute := replicationcontroller.NewReplicationControllerRoute()
	rcRoute.RegisterSubRouter(apiGroup)

	// 注册CRD路由
	crdRoute := crd.NewCrdRoute()
	crdRoute.RegisterSubRouter(apiGroup)

	// 注册VPA路由
	vpaRoute := vpa.NewVpaRoute()
	vpaRoute.RegisterSubRouter(apiGroup)

	// 注册HPA路由
	hpaRoute := hpa.NewHpaRoute()
	hpaRoute.RegisterSubRouter(apiGroup)

	// 注册Operator路由
	operatorRoute := operator.NewOperatorRoute()
	operatorRoute.RegisterSubRouter(apiGroup)
}
