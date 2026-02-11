package hpa

import (
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HPAController struct{}

func NewHPAController() *HPAController {
	return &HPAController{}
}

func (c *HPAController) GetHPAList(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
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

	var list *autoscalingv2.HorizontalPodAutoscalerList
	var err error

	if namespace == "all" {
		list, err = client.AutoscalingV2().HorizontalPodAutoscalers("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取HPA列表失败: " + err.Error())
		return
	}

	hpaList := make([]gin.H, 0)
	for _, item := range list.Items {
		hpaList = append(hpaList, c.convertHPAToListItem(item))
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "hpaList", hpaList)
}

func (c *HPAController) GetHPADetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
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

	hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取HPA详情失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "hpaDetail", hpa)
}

func (c *HPAController) DeleteHPA(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
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

	err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除HPA失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("HPA删除成功")
}

func (c *HPAController) convertHPAToListItem(hpa autoscalingv2.HorizontalPodAutoscaler) gin.H {
	var minReplicas int32 = 0
	if hpa.Spec.MinReplicas != nil {
		minReplicas = *hpa.Spec.MinReplicas
	}
	return gin.H{
		"name":      hpa.Name,
		"namespace": hpa.Namespace,
		"target":    hpa.Spec.ScaleTargetRef.Name,
		"min":       minReplicas,
		"max":       hpa.Spec.MaxReplicas,
		"replicas":  hpa.Status.CurrentReplicas,
		"age":       hpa.CreationTimestamp.Unix(),
	}
}
