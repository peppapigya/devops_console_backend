package replicationcontroller

import (
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReplicationControllerController struct{}

func NewReplicationControllerController() *ReplicationControllerController {
	return &ReplicationControllerController{}
}

func (c *ReplicationControllerController) GetRCList(ctx *gin.Context) {
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

	var list *corev1.ReplicationControllerList
	var err error

	if namespace == "all" {
		list, err = client.CoreV1().ReplicationControllers("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.CoreV1().ReplicationControllers(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取ReplicationController列表失败: " + err.Error())
		return
	}

	rcList := make([]gin.H, 0)
	for _, item := range list.Items {
		rcList = append(rcList, c.convertRCToListItem(item))
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "rcList", rcList)
}

func (c *ReplicationControllerController) GetRCDetail(ctx *gin.Context) {
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

	rc, err := client.CoreV1().ReplicationControllers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取ReplicationController详情失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "rcDetail", rc)
}

func (c *ReplicationControllerController) DeleteRC(ctx *gin.Context) {
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

	err := client.CoreV1().ReplicationControllers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除ReplicationController失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("ReplicationController删除成功")
}

func (c *ReplicationControllerController) convertRCToListItem(rc corev1.ReplicationController) gin.H {
	var replicas int32 = 0
	if rc.Spec.Replicas != nil {
		replicas = *rc.Spec.Replicas
	}
	return gin.H{
		"name":      rc.Name,
		"namespace": rc.Namespace,
		"desired":   replicas,
		"current":   rc.Status.Replicas,
		"ready":     rc.Status.ReadyReplicas,
		"age":       rc.CreationTimestamp.Unix(),
		"images":    getImages(rc.Spec.Template.Spec.Containers),
	}
}

func getImages(containers []corev1.Container) []string {
	images := make([]string, 0, len(containers))
	for _, c := range containers {
		images = append(images, c.Image)
	}
	return images
}
