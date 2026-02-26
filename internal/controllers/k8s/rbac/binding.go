package rbac

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops-console-backend/pkg/configs"
)

// RoleBindingController RoleBinding控制器
type RoleBindingController struct{}

// NewRoleBindingController 创建RoleBinding控制器实例
func NewRoleBindingController() *RoleBindingController {
	return &RoleBindingController{}
}

// GetRoleBindingList 获取RoleBinding列表
func (c *RoleBindingController) GetRoleBindingList(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	var list *rbacv1.RoleBindingList
	var err error
	if namespace == "all" || namespace == "" {
		list, err = client.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("获取RoleBinding列表失败: " + err.Error())
		return
	}

	roleBindingList := make([]k8s.RoleBindingListItem, 0, len(list.Items))
	for _, item := range list.Items {
		roleBindingList = append(roleBindingList, k8s.RoleBindingListItem{
			Name:          item.Name,
			Namespace:     item.Namespace,
			RoleRefKind:   item.RoleRef.Kind,
			RoleRefName:   item.RoleRef.Name,
			SubjectsCount: len(item.Subjects),
			Age:           item.CreationTimestamp.Unix(),
		})
	}
	utils.NewResponseHelper(ctx).SuccessWithData("success", "roleBindingList", roleBindingList)
}

// GetRoleBindingDetail 获取RoleBinding详情
func (c *RoleBindingController) GetRoleBindingDetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	rb, err := client.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).NotFound("RoleBinding 不存在")
		return
	}
	utils.NewResponseHelper(ctx).SuccessWithData("success", "roleBindingDetail", rb)
}

// CreateRoleBinding 创建RoleBinding
func (c *RoleBindingController) CreateRoleBinding(ctx *gin.Context) {
	var req k8s.RoleBindingCreateRequest
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

	var rb *rbacv1.RoleBinding
	var err error

	if req.YAML != "" {
		rb, err = parseYAMLToRoleBinding(req.YAML)
		if err != nil {
			utils.NewResponseHelper(ctx).BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		rb = &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     req.RoleRef.Kind,
				Name:     req.RoleRef.Name,
			},
			Subjects: convertSubjects(req.Subjects),
		}
	}

	_, err = client.RbacV1().RoleBindings(rb.Namespace).Create(ctx, rb, metav1.CreateOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("创建RoleBinding失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("RoleBinding创建成功")
}

// UpdateRoleBinding 更新RoleBinding
func (c *RoleBindingController) UpdateRoleBinding(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	var req k8s.RoleBindingUpdateRequest
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

	var rb *rbacv1.RoleBinding
	var err error

	if req.YAML != "" {
		rb, err = parseYAMLToRoleBinding(req.YAML)
		if err != nil {
			utils.NewResponseHelper(ctx).BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		existing, err := client.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			utils.NewResponseHelper(ctx).NotFound("RoleBinding 不存在")
			return
		}
		existing.Subjects = convertSubjects(req.Subjects)
		rb = existing
	}

	_, err = client.RbacV1().RoleBindings(namespace).Update(ctx, rb, metav1.UpdateOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("更新RoleBinding失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("RoleBinding更新成功")
}

// DeleteRoleBinding 删除RoleBinding
func (c *RoleBindingController) DeleteRoleBinding(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	err := client.RbacV1().RoleBindings(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("删除RoleBinding失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("RoleBinding删除成功")
}

// ==================== ClusterRoleBinding ====================

// ClusterRoleBindingController ClusterRoleBinding控制器
type ClusterRoleBindingController struct{}

// NewClusterRoleBindingController 创建ClusterRoleBinding控制器实例
func NewClusterRoleBindingController() *ClusterRoleBindingController {
	return &ClusterRoleBindingController{}
}

// GetClusterRoleBindingList 获取ClusterRoleBinding列表
func (c *ClusterRoleBindingController) GetClusterRoleBindingList(ctx *gin.Context) {
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	list, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("获取ClusterRoleBinding列表失败: " + err.Error())
		return
	}

	crbList := make([]k8s.ClusterRoleBindingListItem, 0, len(list.Items))
	for _, item := range list.Items {
		crbList = append(crbList, k8s.ClusterRoleBindingListItem{
			Name:          item.Name,
			RoleRefName:   item.RoleRef.Name,
			SubjectsCount: len(item.Subjects),
			Age:           item.CreationTimestamp.Unix(),
		})
	}
	utils.NewResponseHelper(ctx).SuccessWithData("success", "clusterRoleBindingList", crbList)
}

// GetClusterRoleBindingDetail 获取ClusterRoleBinding详情
func (c *ClusterRoleBindingController) GetClusterRoleBindingDetail(ctx *gin.Context) {
	name := ctx.Param("name")
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	crb, err := client.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).NotFound("ClusterRoleBinding 不存在")
		return
	}
	utils.NewResponseHelper(ctx).SuccessWithData("success", "clusterRoleBindingDetail", crb)
}

// CreateClusterRoleBinding 创建ClusterRoleBinding
func (c *ClusterRoleBindingController) CreateClusterRoleBinding(ctx *gin.Context) {
	var req k8s.ClusterRoleBindingCreateRequest
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

	var crb *rbacv1.ClusterRoleBinding
	var err error

	if req.YAML != "" {
		crb, err = parseYAMLToClusterRoleBinding(req.YAML)
		if err != nil {
			utils.NewResponseHelper(ctx).BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		crb = &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: req.Name},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     req.RoleRef.Kind,
				Name:     req.RoleRef.Name,
			},
			Subjects: convertSubjects(req.Subjects),
		}
	}

	_, err = client.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("创建ClusterRoleBinding失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("ClusterRoleBinding创建成功")
}

// UpdateClusterRoleBinding 更新ClusterRoleBinding
func (c *ClusterRoleBindingController) UpdateClusterRoleBinding(ctx *gin.Context) {
	name := ctx.Param("name")

	var req k8s.ClusterRoleBindingUpdateRequest
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

	var crb *rbacv1.ClusterRoleBinding
	var err error

	if req.YAML != "" {
		crb, err = parseYAMLToClusterRoleBinding(req.YAML)
		if err != nil {
			utils.NewResponseHelper(ctx).BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		existing, err := client.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			utils.NewResponseHelper(ctx).NotFound("ClusterRoleBinding 不存在")
			return
		}
		existing.Subjects = convertSubjects(req.Subjects)
		crb = existing
	}

	_, err = client.RbacV1().ClusterRoleBindings().Update(ctx, crb, metav1.UpdateOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("更新ClusterRoleBinding失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("ClusterRoleBinding更新成功")
}

// DeleteClusterRoleBinding 删除ClusterRoleBinding
func (c *ClusterRoleBindingController) DeleteClusterRoleBinding(ctx *gin.Context) {
	name := ctx.Param("name")
	instanceID := getInstanceID(ctx)

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		utils.NewResponseHelper(ctx).InternalError("K8s客户端未初始化")
		return
	}

	err := client.RbacV1().ClusterRoleBindings().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		utils.NewResponseHelper(ctx).InternalError("删除ClusterRoleBinding失败: " + err.Error())
		return
	}
	utils.NewResponseHelper(ctx).Success("ClusterRoleBinding删除成功")
}

// ==================== 工具函数 ====================

func convertSubjects(subjects []k8s.RoleBindingSubject) []rbacv1.Subject {
	result := make([]rbacv1.Subject, 0, len(subjects))
	for _, s := range subjects {
		subj := rbacv1.Subject{
			Kind: s.Kind,
			Name: s.Name,
		}
		if s.Kind == "ServiceAccount" {
			subj.Namespace = s.Namespace
			subj.APIGroup = ""
		} else {
			subj.APIGroup = "rbac.authorization.k8s.io"
		}
		result = append(result, subj)
	}
	return result
}

func parseYAMLToRoleBinding(yamlContent string) (*rbacv1.RoleBinding, error) {
	var rb rbacv1.RoleBinding
	if err := yaml.Unmarshal([]byte(yamlContent), &rb); err != nil {
		return nil, err
	}
	return &rb, nil
}

func parseYAMLToClusterRoleBinding(yamlContent string) (*rbacv1.ClusterRoleBinding, error) {
	var crb rbacv1.ClusterRoleBinding
	if err := yaml.Unmarshal([]byte(yamlContent), &crb); err != nil {
		return nil, err
	}
	return &crb, nil
}
