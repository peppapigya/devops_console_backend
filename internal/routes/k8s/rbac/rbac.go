package rbac

import (
	"devops-console-backend/internal/controllers/k8s/rbac"

	"github.com/gin-gonic/gin"
)

// RbacRoute RBAC路由
type RbacRoute struct {
	roleController               *rbac.RoleController
	clusterRoleController        *rbac.ClusterRoleController
	roleBindingController        *rbac.RoleBindingController
	clusterRoleBindingController *rbac.ClusterRoleBindingController
}

// NewRbacRoute 创建RBAC路由实例
func NewRbacRoute() *RbacRoute {
	return &RbacRoute{
		roleController:               rbac.NewRoleController(),
		clusterRoleController:        rbac.NewClusterRoleController(),
		roleBindingController:        rbac.NewRoleBindingController(),
		clusterRoleBindingController: rbac.NewClusterRoleBindingController(),
	}
}

// RegisterSubRouter 注册RBAC子路由
func (r *RbacRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	// Role（命名空间级）
	roleGroup := apiGroup.Group("/k8s/role")
	{
		roleGroup.GET("/list/:namespace", r.roleController.GetRoleList)
		roleGroup.GET("/list/all", r.roleController.GetRoleList)
		roleGroup.GET("/detail/:namespace/:name", r.roleController.GetRoleDetail)
		roleGroup.POST("/create", r.roleController.CreateRole)
		roleGroup.PUT("/update/:namespace/:name", r.roleController.UpdateRole)
		roleGroup.DELETE("/delete/:namespace/:name", r.roleController.DeleteRole)
	}

	// ClusterRole（集群级）
	clusterRoleGroup := apiGroup.Group("/k8s/clusterrole")
	{
		clusterRoleGroup.GET("/list", r.clusterRoleController.GetClusterRoleList)
		clusterRoleGroup.GET("/detail/:name", r.clusterRoleController.GetClusterRoleDetail)
		clusterRoleGroup.POST("/create", r.clusterRoleController.CreateClusterRole)
		clusterRoleGroup.PUT("/update/:name", r.clusterRoleController.UpdateClusterRole)
		clusterRoleGroup.DELETE("/delete/:name", r.clusterRoleController.DeleteClusterRole)
	}

	// RoleBinding（命名空间级）
	roleBindingGroup := apiGroup.Group("/k8s/rolebinding")
	{
		roleBindingGroup.GET("/list/:namespace", r.roleBindingController.GetRoleBindingList)
		roleBindingGroup.GET("/list/all", r.roleBindingController.GetRoleBindingList)
		roleBindingGroup.GET("/detail/:namespace/:name", r.roleBindingController.GetRoleBindingDetail)
		roleBindingGroup.POST("/create", r.roleBindingController.CreateRoleBinding)
		roleBindingGroup.PUT("/update/:namespace/:name", r.roleBindingController.UpdateRoleBinding)
		roleBindingGroup.DELETE("/delete/:namespace/:name", r.roleBindingController.DeleteRoleBinding)
	}

	// ClusterRoleBinding（集群级）
	clusterRoleBindingGroup := apiGroup.Group("/k8s/clusterrolebinding")
	{
		clusterRoleBindingGroup.GET("/list", r.clusterRoleBindingController.GetClusterRoleBindingList)
		clusterRoleBindingGroup.GET("/detail/:name", r.clusterRoleBindingController.GetClusterRoleBindingDetail)
		clusterRoleBindingGroup.POST("/create", r.clusterRoleBindingController.CreateClusterRoleBinding)
		clusterRoleBindingGroup.PUT("/update/:name", r.clusterRoleBindingController.UpdateClusterRoleBinding)
		clusterRoleBindingGroup.DELETE("/delete/:name", r.clusterRoleBindingController.DeleteClusterRoleBinding)
	}
}
