package operator

import (
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// OperatorController 主要管理 Subscription 和 CSV
type OperatorController struct{}

func NewOperatorController() *OperatorController {
	return &OperatorController{}
}

// Subscription GVR
var subGVR = schema.GroupVersionResource{
	Group:    "operators.coreos.com",
	Version:  "v1alpha1",
	Resource: "subscriptions",
}

func (c *OperatorController) GetSubscriptionList(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetDynamicClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s动态客户端未初始化")
		return
	}

	var list interface{}
	var err error

	if namespace == "all" {
		list, err = client.Resource(subGVR).List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.Resource(subGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Subscription列表失败(请确认OLM是否已安装): " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "subscriptionList", list)
}

func (c *OperatorController) GetSubscriptionDetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetDynamicClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s动态客户端未初始化")
		return
	}

	sub, err := client.Resource(subGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Subscription详情失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "subscriptionDetail", sub)
}

func (c *OperatorController) DeleteSubscription(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetDynamicClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s动态客户端未初始化")
		return
	}

	err := client.Resource(subGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除Subscription失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Subscription删除成功")
}
