package daemonset

import (
	"devops-console-backend/config"
	"devops-console-backend/models/k8s"
	"devops-console-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSetController DaemonSet控制器
type DaemonSetController struct{}

// NewDaemonSetController 创建DaemonSet控制器实例
func NewDaemonSetController() *DaemonSetController {
	return &DaemonSetController{}
}

// GetDaemonSetDetail 获取DaemonSet详情
func (c *DaemonSetController) GetDaemonSetDetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	daemonSetName := ctx.Param("daemonSetName")
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

	daemonSetDetail, err := client.AppsV1().DaemonSets(namespace).Get(ctx, daemonSetName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("DaemonSet不存在")
		return
	}

	// 转换Conditions为interface{}
	conditions := make([]interface{}, len(daemonSetDetail.Status.Conditions))
	for i, condition := range daemonSetDetail.Status.Conditions {
		conditions[i] = condition
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "daemonSetDetail", k8s.DaemonSetDetail{
		Name:       daemonSetDetail.Name,
		Namespace:  daemonSetDetail.Namespace,
		Current:    daemonSetDetail.Status.CurrentNumberScheduled,
		Desired:    daemonSetDetail.Status.DesiredNumberScheduled,
		Ready:      daemonSetDetail.Status.NumberReady,
		Available:  daemonSetDetail.Status.NumberAvailable,
		Conditions: conditions,
		Labels:     daemonSetDetail.Labels,
		Age:        daemonSetDetail.CreationTimestamp.Unix(),
	})
}

// GetDaemonSetList 获取DaemonSet列表
func (c *DaemonSetController) GetDaemonSetList(ctx *gin.Context) {
	namespace := ctx.Param("namespace")

	// 如果 namespace 为 "all"，则使用空字符串获取所有命名空间的资源
	if namespace == "all" {
		namespace = ""
	}

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

	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取DaemonSet列表失败")
		return
	}

	// 简化返回数据，只返回关键信息
	var simplifiedList []k8s.DaemonSetListItem
	for _, daemonSet := range daemonSetList.Items {
		simplifiedList = append(simplifiedList, k8s.DaemonSetListItem{
			Name:      daemonSet.Name,
			Namespace: daemonSet.Namespace,
			Current:   daemonSet.Status.CurrentNumberScheduled,
			Desired:   daemonSet.Status.DesiredNumberScheduled,
			Ready:     daemonSet.Status.NumberReady,
			Available: daemonSet.Status.NumberAvailable,
			Created:   daemonSet.CreationTimestamp.Time,
		})
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "daemonSetList", simplifiedList)
}

// CreateDaemonSet 创建DaemonSet
func (c *DaemonSetController) CreateDaemonSet(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	var daemonSetReq k8s.DaemonSetCreateRequest

	if err := ctx.ShouldBindJSON(&daemonSetReq); err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

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

	daemonSet := c.convertCreateRequestToK8sDaemonSet(namespace, daemonSetReq)
	_, err := client.AppsV1().DaemonSets(namespace).Create(ctx, daemonSet, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建DaemonSet失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("DaemonSet创建成功", "data", map[string]interface{}{
		"name":      daemonSetReq.Name,
		"namespace": namespace,
	})
}

// UpdateDaemonSet 更新DaemonSet
func (c *DaemonSetController) UpdateDaemonSet(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	daemonSetName := ctx.Param("daemonSetName")
	var updateReq k8s.DaemonSetUpdateRequest

	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

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

	// 获取现有的DaemonSet
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(ctx, daemonSetName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("DaemonSet不存在")
		return
	}

	// 更新镜像版本
	if len(daemonSet.Spec.Template.Spec.Containers) > 0 {
		daemonSet.Spec.Template.Spec.Containers[0].Image = updateReq.Image
	}

	// 添加更新注释
	if daemonSet.Annotations == nil {
		daemonSet.Annotations = make(map[string]string)
	}
	daemonSet.Annotations["updated-by"] = "devops-console"
	daemonSet.Annotations["updated-at"] = metav1.Now().String()

	_, err = client.AppsV1().DaemonSets(namespace).Update(ctx, daemonSet, metav1.UpdateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("更新DaemonSet失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("DaemonSet更新成功", "data", map[string]interface{}{
		"name":      daemonSet.Name,
		"namespace": daemonSet.Namespace,
		"image":     updateReq.Image,
	})
}

// DeleteDaemonSet 删除DaemonSet
func (c *DaemonSetController) DeleteDaemonSet(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	daemonSetName := ctx.Param("daemonSetName")

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

	deletePolicy := metav1.DeletePropagationForeground
	err := client.AppsV1().DaemonSets(namespace).Delete(ctx, daemonSetName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除DaemonSet失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("DaemonSet删除成功", "data", map[string]interface{}{
		"name":      daemonSetName,
		"namespace": namespace,
	})
}

// convertCreateRequestToK8sDaemonSet 转换创建请求为K8s DaemonSet
func (c *DaemonSetController) convertCreateRequestToK8sDaemonSet(namespace string, req k8s.DaemonSetCreateRequest) *appsv1.DaemonSet {
	if req.Labels == nil {
		req.Labels = map[string]string{
			"app": req.Name,
		}
	}

	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			Labels:    req.Labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: req.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: req.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  req.Name,
							Image: req.Image,
							Ports: []corev1.ContainerPort{},
						},
					},
				},
			},
		},
	}

	// 如果指定了端口，添加端口配置
	if req.Port > 0 {
		daemonSet.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: req.Port,
				Protocol:      corev1.ProtocolTCP,
			},
		}
	}

	return daemonSet
}
