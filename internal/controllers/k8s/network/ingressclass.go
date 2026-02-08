package network

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressClassController IngressClass控制器
type IngressClassController struct{}

// NewIngressClassController 创建IngressClass控制器实例
func NewIngressClassController() *IngressClassController {
	return &IngressClassController{}
}

// GetIngressClassList 获取IngressClass列表
func (c *IngressClassController) GetIngressClassList(ctx *gin.Context) {
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	var listOptions metav1.ListOptions
	list, err := client.NetworkingV1().IngressClasses().List(ctx, listOptions)

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取IngressClass列表失败: " + err.Error())
		return
	}

	ingressClassList := make([]k8s.IngressClassListItem, 0)
	for _, item := range list.Items {
		isDefault := false
		if item.Annotations != nil {
			if val, ok := item.Annotations["ingressclass.kubernetes.io/is-default-class"]; ok && val == "true" {
				isDefault = true
			}
		}

		var parameters *string
		if item.Spec.Parameters != nil {
			paramStr := item.Spec.Parameters.Kind + "/" + item.Spec.Parameters.Name
			parameters = &paramStr
		}

		ingressClassList = append(ingressClassList, k8s.IngressClassListItem{
			Name:       item.Name,
			Controller: item.Spec.Controller,
			IsDefault:  isDefault,
			Parameters: parameters,
			Age:        item.CreationTimestamp.Unix(),
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "ingressClassList", ingressClassList)
}

// GetIngressClassDetail 获取IngressClass详情
func (c *IngressClassController) GetIngressClassDetail(ctx *gin.Context) {
	name := ctx.Param("name")
	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	ingressClass, err := client.NetworkingV1().IngressClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("IngressClass 不存在")
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "ingressClassDetail", ingressClass)
}
