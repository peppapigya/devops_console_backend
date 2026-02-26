package rbac

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleController Role控制器
type RoleController struct{}

// NewRoleController 创建Role控制器实例
func NewRoleController() *RoleController {
	return &RoleController{}
}

// getInstanceID 提取 instance_id 查询参数
func getInstanceID(ctx *gin.Context) uint {
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}
	return instanceID
}

// GetRoleList 获取Role列表
func (c *RoleController) GetRoleList(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	var list *rbacv1.RoleList
	var err error
	if namespace == "all" || namespace == "" {
		list, err = client.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("获取Role列表失败: " + err.Error())
		return
	}

	roleList := make([]k8s.RoleListItem, 0, len(list.Items))
	for _, item := range list.Items {
		roleList = append(roleList, k8s.RoleListItem{
			Name:       item.Name,
			Namespace:  item.Namespace,
			RulesCount: len(item.Rules),
			Age:        item.CreationTimestamp.Unix(),
		})
	}
	utils.NewResponseHelper(ctx).SuccessWithData("success", "roleList", roleList)
}

// GetRoleDetail 获取Role详情
func (c *RoleController) GetRoleDetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	role, err := client.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).NotFound("Role 不存在")
		return
	}
	utils.NewResponseHelper(ctx).SuccessWithData("success", "roleDetail", role)
}

// CreateRole 创建Role
func (c *RoleController) CreateRole(ctx *gin.Context) {
	var req k8s.RoleCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.NewResponseHelper(ctx).BadRequest("请求参数错误: " + err.Error())
		return
	}
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	var role *rbacv1.Role
	var err error

	if req.YAML != "" {
		role, err = parseYAMLToRole(req.YAML)
		if err != nil {
			utils.NewResponseHelper(ctx).BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		role = &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
			Rules: convertPolicyRules(req.Rules),
		}
	}

	_, err = client.RbacV1().Roles(role.Namespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("创建Role失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("Role创建成功")
}

// UpdateRole 更新Role
func (c *RoleController) UpdateRole(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	var req k8s.RoleUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.NewResponseHelper(ctx).BadRequest("请求参数错误: " + err.Error())
		return
	}
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	var role *rbacv1.Role
	var err error

	if req.YAML != "" {
		role, err = parseYAMLToRole(req.YAML)
		if err != nil {
			utils.NewResponseHelper(ctx).BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		existing, err := client.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			utils.NewResponseHelper(ctx).NotFound("Role 不存在")
			return
		}
		existing.Rules = convertPolicyRules(req.Rules)
		role = existing
	}

	_, err = client.RbacV1().Roles(namespace).Update(ctx, role, metav1.UpdateOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("更新Role失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("Role更新成功")
}

// DeleteRole 删除Role
func (c *RoleController) DeleteRole(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	err := client.RbacV1().Roles(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("删除Role失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("Role删除成功")
}

// ==================== ClusterRole ====================

// ClusterRoleController ClusterRole控制器
type ClusterRoleController struct{}

// NewClusterRoleController 创建ClusterRole控制器实例
func NewClusterRoleController() *ClusterRoleController {
	return &ClusterRoleController{}
}

// GetClusterRoleList 获取ClusterRole列表
func (c *ClusterRoleController) GetClusterRoleList(ctx *gin.Context) {
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	list, err := client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("获取ClusterRole列表失败: " + err.Error())
		return
	}

	clusterRoleList := make([]k8s.ClusterRoleListItem, 0, len(list.Items))
	for _, item := range list.Items {
		clusterRoleList = append(clusterRoleList, k8s.ClusterRoleListItem{
			Name:            item.Name,
			RulesCount:      len(item.Rules),
			AggregationRule: item.AggregationRule != nil,
			Age:             item.CreationTimestamp.Unix(),
		})
	}
	utils.NewResponseHelper(ctx).SuccessWithData("success", "clusterRoleList", clusterRoleList)
}

// GetClusterRoleDetail 获取ClusterRole详情
func (c *ClusterRoleController) GetClusterRoleDetail(ctx *gin.Context) {
	name := ctx.Param("name")
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	cr, err := client.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).NotFound("ClusterRole 不存在")
		return
	}
	utils.NewResponseHelper(ctx).SuccessWithData("success", "clusterRoleDetail", cr)
}

// CreateClusterRole 创建ClusterRole
func (c *ClusterRoleController) CreateClusterRole(ctx *gin.Context) {
	var req k8s.ClusterRoleCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.NewResponseHelper(ctx).BadRequest("请求参数错误: " + err.Error())
		return
	}
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	var cr *rbacv1.ClusterRole
	var err error

	if req.YAML != "" {
		cr, err = parseYAMLToClusterRole(req.YAML)
		if err != nil {
			utils.NewResponseHelper(ctx).BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		cr = &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: req.Name},
			Rules:      convertPolicyRules(req.Rules),
		}
	}

	_, err = client.RbacV1().ClusterRoles().Create(ctx, cr, metav1.CreateOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("创建ClusterRole失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("ClusterRole创建成功")
}

// UpdateClusterRole 更新ClusterRole
func (c *ClusterRoleController) UpdateClusterRole(ctx *gin.Context) {
	name := ctx.Param("name")

	var req k8s.ClusterRoleUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.NewResponseHelper(ctx).BadRequest("请求参数错误: " + err.Error())
		return
	}
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	var cr *rbacv1.ClusterRole
	var err error

	if req.YAML != "" {
		cr, err = parseYAMLToClusterRole(req.YAML)
		if err != nil {
			utils.NewResponseHelper(ctx).BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		existing, err := client.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			utils.NewResponseHelper(ctx).NotFound("ClusterRole 不存在")
			return
		}
		existing.Rules = convertPolicyRules(req.Rules)
		cr = existing
	}

	_, err = client.RbacV1().ClusterRoles().Update(ctx, cr, metav1.UpdateOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("更新ClusterRole失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("ClusterRole更新成功")
}

// DeleteClusterRole 删除ClusterRole
func (c *ClusterRoleController) DeleteClusterRole(ctx *gin.Context) {
	name := ctx.Param("name")
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	err := client.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("删除ClusterRole失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("ClusterRole删除成功")
}

// ==================== 工具函数 ====================

// convertPolicyRules 将请求中的规则转换为K8s RBAC规则
func convertPolicyRules(rules []k8s.PolicyRule) []rbacv1.PolicyRule {
	result := make([]rbacv1.PolicyRule, 0, len(rules))
	for _, r := range rules {
		result = append(result, rbacv1.PolicyRule{
			APIGroups: r.APIGroups,
			Resources: r.Resources,
			Verbs:     r.Verbs,
		})
	}
	return result
}

func parseYAMLToRole(yamlContent string) (*rbacv1.Role, error) {
	var role rbacv1.Role
	if err := yaml.Unmarshal([]byte(yamlContent), &role); err != nil {
		return nil, err
	}
	return &role, nil
}

func parseYAMLToClusterRole(yamlContent string) (*rbacv1.ClusterRole, error) {
	var cr rbacv1.ClusterRole
	if err := yaml.Unmarshal([]byte(yamlContent), &cr); err != nil {
		return nil, err
	}
	return &cr, nil
}
