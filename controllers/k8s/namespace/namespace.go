package namespace

import (
	"devops-console-backend/config"
	"devops-console-backend/models/k8s"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceController Namespace控制器
type NamespaceController struct{}

// NewNamespaceController 创建Namespace控制器实例
func NewNamespaceController() *NamespaceController {
	return &NamespaceController{}
}

// CreateNamespace 创建Namespace
func (c *NamespaceController) CreateNamespace(ctx *gin.Context) {
	namespaceName := ctx.Param("namespace")
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	_, err := client.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建Namespace失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("创建Namespace成功")
}

// DeleteNamespace 删除Namespace
func (c *NamespaceController) DeleteNamespace(ctx *gin.Context) {
	namespaceName := ctx.Param("namespace")
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	err := client.CoreV1().Namespaces().Delete(ctx, namespaceName, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除Namespace失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("删除Namespace成功")
}

// GetNamespaceList 获取Namespace列表
func (c *NamespaceController) GetNamespaceList(ctx *gin.Context) {
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	list, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("查询Namespace列表失败")
		return
	}

	namespaceList := []k8s.NamespaceListItem{}
	for _, ns := range list.Items {
		namespaceInfo := k8s.NamespaceListItem{
			Name:              ns.Name,
			Status:            string(ns.Status.Phase),
			CreationTimestamp: ns.CreationTimestamp.Time,
			Labels:            ns.Labels,
			Annotations:       ns.Annotations,
			Age:               ns.CreationTimestamp.Unix(),
		}
		namespaceList = append(namespaceList, namespaceInfo)
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "namespaceList", namespaceList)
}
