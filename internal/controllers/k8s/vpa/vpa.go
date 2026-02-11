package vpa

import (
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type VPAController struct{}

func NewVPAController() *VPAController {
	return &VPAController{}
}

var vpaGVR = schema.GroupVersionResource{
	Group:    "autoscaling.k8s.io",
	Version:  "v1",
	Resource: "verticalpodautoscalers",
}

func (c *VPAController) GetVPAList(ctx *gin.Context) {
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

	// VPA 通常是 autoscaling.k8s.io/v1
	// 如果没有安装 VPA CRD，这里会报错
	if namespace == "all" {
		list, err = client.Resource(vpaGVR).List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.Resource(vpaGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		// 可能是CRD不存在
		helper.InternalError("获取VPA列表失败(请确认VPA是否已安装): " + err.Error())
		return
	}

	// 转换可以不做，直接返回unstructured列表
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "vpaList", list)
}

func (c *VPAController) GetVPADetail(ctx *gin.Context) {
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

	vpa, err := client.Resource(vpaGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取VPA详情失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "vpaDetail", vpa)
}

func (c *VPAController) DeleteVPA(ctx *gin.Context) {
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

	err := client.Resource(vpaGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除VPA失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("VPA删除成功")
}
