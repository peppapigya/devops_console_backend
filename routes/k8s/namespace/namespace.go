package namespace

import (
	"devops-console-backend/controllers/k8s/namespace"
	"github.com/gin-gonic/gin"
)

// NamespaceRoute Namespace路由
type NamespaceRoute struct {
	controller *namespace.NamespaceController
}

// NewNamespaceRoute 创建Namespace路由实例
func NewNamespaceRoute() *NamespaceRoute {
	return &NamespaceRoute{
		controller: namespace.NewNamespaceController(),
	}
}

// RegisterSubRouter 注册子路由
func (r *NamespaceRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	namespaceGroup := apiGroup.Group("/k8s/namespace")
	{
		namespaceGroup.POST("/create/:namespace", r.controller.CreateNamespace)
		namespaceGroup.DELETE("/delete/:namespace", r.controller.DeleteNamespace)
		namespaceGroup.GET("/list", r.controller.GetNamespaceList)
	}
}