package event

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventController Event控制器
type EventController struct{}

// NewEventController 创建Event控制器实例
func NewEventController() *EventController {
	return &EventController{}
}

// GetEventList 获取Event列表
func (c *EventController) GetEventList(ctx *gin.Context) {
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

	var listOptions metav1.ListOptions
	var list *corev1.EventList
	var err error

	if namespace == "all" || namespace == "" {
		list, err = client.CoreV1().Events("").List(ctx, listOptions)
	} else {
		list, err = client.CoreV1().Events(namespace).List(ctx, listOptions)
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Event列表失败: " + err.Error())
		return
	}

	eventList := make([]k8s.EventListItem, 0)
	for _, item := range list.Items {
		eventList = append(eventList, k8s.EventListItem{
			Name:           item.Name,
			Namespace:      item.Namespace,
			Type:           item.Type,
			Reason:         item.Reason,
			Message:        item.Message,
			InvolvedObject: item.InvolvedObject.Name,
			InvolvedKind:   item.InvolvedObject.Kind,
			Source:         item.Source.Component,
			Count:          item.Count,
			FirstTimestamp: item.FirstTimestamp.Unix(),
			LastTimestamp:  item.LastTimestamp.Unix(),
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "eventList", eventList)
}
