package statefulset

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/logs"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSetController StatefulSet控制器
type StatefulSetController struct{}

// NewStatefulSetController 创建StatefulSet控制器实例
func NewStatefulSetController() *StatefulSetController {
	return &StatefulSetController{}
}

// GetStatefulSetDetail 获取StatefulSet详情
func (c *StatefulSetController) GetStatefulSetDetail(ctx *gin.Context) {
	logData := map[string]interface{}{
		"namespace":       ctx.Param("namespace"),
		"statefulSetName": ctx.Param("name"),
	}
	logs.Debug(logData, "获取StatefulSet详情")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	stsDetail, err := client.AppsV1().StatefulSets(logData["namespace"].(string)).Get(ctx, logData["statefulSetName"].(string), metav1.GetOptions{})
	if err != nil {
		logs.Error(logData, "获取StatefulSet失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("StatefulSet不存在")
		return
	}

	// 转换Conditions为interface{}
	conditions := make([]interface{}, len(stsDetail.Status.Conditions))
	for i, condition := range stsDetail.Status.Conditions {
		conditions[i] = condition
	}

	// 提取容器信息
	var containers []k8s.ContainerInfo
	for _, c := range stsDetail.Spec.Template.Spec.Containers {
		containers = append(containers, k8s.ContainerInfo{
			Name:  c.Name,
			Image: c.Image,
		})
	}

	logs.Info(logData, "获取StatefulSet详情成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "statefulSetDetail", k8s.StatefulSetDetail{
		Name:       stsDetail.Name,
		Namespace:  stsDetail.Namespace,
		Replicas:   stsDetail.Status.Replicas,
		Ready:      stsDetail.Status.ReadyReplicas,
		Conditions: conditions,
		Labels:     stsDetail.Labels,
		Selector:   stsDetail.Spec.Selector.MatchLabels,
		Age:        stsDetail.CreationTimestamp.Unix(),
		Containers: containers,
	})
}

// GetStatefulSetList 获取StatefulSet列表
func (c *StatefulSetController) GetStatefulSetList(ctx *gin.Context) {
	namespace := ctx.Query("namespace")

	// 如果 namespace 为 "all"，则使用空字符串获取所有命名空间的资源
	if namespace == "all" || namespace == "" {
		namespace = ""
	}

	logData := map[string]interface{}{"namespace": namespace}
	logs.Debug(logData, "获取 StatefulSet列表")

	instanceIDStr := ctx.Query("instanceId")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	stsList, err := client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logs.Error(logData, "获取StatefulSet列表失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取 StatefulSet列表失败")
		return
	}

	// 简化返回数据，只返回关键信息
	var simplifiedList []k8s.StatefulSetListItem
	for _, sts := range stsList.Items {
		// 提取第一个容器的镜像和资源配额
		image := ""
		resources := k8s.ResourceInfo{
			CPURequest:    "0",
			CPULimit:      "0",
			MemoryRequest: "0",
			MemoryLimit:   "0",
		}

		if len(sts.Spec.Template.Spec.Containers) > 0 {
			c := sts.Spec.Template.Spec.Containers[0]
			image = c.Image

			// 获取资源限制
			if c.Resources.Requests != nil {
				resources.CPURequest = c.Resources.Requests.Cpu().String()
				resources.MemoryRequest = c.Resources.Requests.Memory().String()
			}
			if c.Resources.Limits != nil {
				resources.CPULimit = c.Resources.Limits.Cpu().String()
				resources.MemoryLimit = c.Resources.Limits.Memory().String()
			}
		}

		var replicas int32 = 0
		if sts.Spec.Replicas != nil {
			replicas = *sts.Spec.Replicas
		}

		simplifiedList = append(simplifiedList, k8s.StatefulSetListItem{
			Name:      sts.Name,
			Namespace: sts.Namespace,
			Replicas:  replicas,
			Ready:     sts.Status.ReadyReplicas,
			Created:   sts.CreationTimestamp.Time,
			Image:     image,
			Resources: resources,
			Labels:    sts.Labels,
		})
	}

	logs.Info(map[string]interface{}{"count": len(simplifiedList), "data": logData}, "获取StatefulSet列表成功")
	helper := utils.NewResponseHelper(ctx)

	// 注意前端期望的数据格式
	helper.SuccessWithData("success", "statefulSetList", simplifiedList)
}

// CreateStatefulSet 创建StatefulSet
func (c *StatefulSetController) CreateStatefulSet(ctx *gin.Context) {
	logData := map[string]interface{}{}
	logs.Debug(logData, "创建 StatefulSet")

	// 这里前端提交可能是一个完整的YAML json或者是基础属性，看前端的 api/k8s/statefulset.js 实际上没传递 namespace in path for create
	// 但看 axios 方法: `/k8s/statefulset/create` data, params: {instanceId}
	// 若前端传的是 yaml 配置，需要额外处理，这里提供基础 JSON 字段的支持
	var createReq k8s.StatefulSetCreateRequest
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		logs.Error(logData, "请求参数绑定失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	// 获取默认或指定的namespace, frontend 如果没传用 default
	namespace := "default"

	instanceIDStr := ctx.Query("instanceId")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	sts := c.convertCreateRequestToK8sStatefulSet(namespace, createReq)
	_, err := client.AppsV1().StatefulSets(namespace).Create(ctx, sts, metav1.CreateOptions{})
	if err != nil {
		logs.Error(map[string]interface{}{"name": createReq.Name, "error": err.Error(), "data": logData}, "创建StatefulSet失败")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建StatefulSet失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{"name": createReq.Name, "data": logData}, "创建StatefulSet成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("StatefulSet创建成功", "data", map[string]interface{}{
		"name":      createReq.Name,
		"namespace": namespace,
	})
}

// UpdateStatefulSet 更新StatefulSet
func (c *StatefulSetController) UpdateStatefulSet(ctx *gin.Context) {
	namespace := ctx.Query("namespace")
	name := ctx.Query("name")

	logData := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	}
	logs.Debug(logData, "更新StatefulSet")

	var updateReq k8s.StatefulSetUpdateRequest
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		logs.Error(logData, "请求参数绑定失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	instanceIDStr := ctx.Query("instanceId")
	instanceID := uint(1) // 默认值
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		logs.Error(logData, "获取StatefulSet失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("StatefulSet不存在")
		return
	}

	// 更新镜像版本
	if len(sts.Spec.Template.Spec.Containers) > 0 {
		sts.Spec.Template.Spec.Containers[0].Image = updateReq.Image
	}

	if sts.Annotations == nil {
		sts.Annotations = make(map[string]string)
	}
	sts.Annotations["updated-by"] = "devops-console"
	sts.Annotations["updated-at"] = metav1.Now().String()

	_, err = client.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		logs.Error(map[string]interface{}{"image": updateReq.Image, "error": err.Error(), "data": logData}, "更新StatefulSet失败")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("更新StatefulSet失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{"image": updateReq.Image, "data": logData}, "更新StatefulSet成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("StatefulSet更新成功", "data", map[string]interface{}{
		"name":      sts.Name,
		"namespace": sts.Namespace,
		"image":     updateReq.Image,
	})
}

// ScaleStatefulSet 扩缩容StatefulSet
func (c *StatefulSetController) ScaleStatefulSet(ctx *gin.Context) {
	namespace := ctx.Query("namespace")
	name := ctx.Query("name")

	logData := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	}
	logs.Info(logData, "接收到StatefulSet扩缩容请求")

	var scaleReq struct {
		Replicas int32 `json:"replicas"`
	}
	if err := ctx.ShouldBindJSON(&scaleReq); err != nil {
		logs.Error(logData, "请求参数绑定失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("请求参数错误: " + err.Error())
		return
	}

	if scaleReq.Replicas < 0 {
		logs.Error(map[string]interface{}{"replicas": scaleReq.Replicas, "data": logData}, "副本数不能为负数")
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("副本数不能为负数")
		return
	}

	instanceIDStr := ctx.Query("instanceId")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		logs.Error(logData, "获取StatefulSet失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		if errors.IsNotFound(err) {
			helper.NotFound("StatefulSet不存在")
		} else {
			helper.InternalError("获取StatefulSet失败: " + err.Error())
		}
		return
	}

	if sts.Spec.Replicas != nil && *sts.Spec.Replicas == scaleReq.Replicas {
		logs.Info(logData, "副本数未改变，跳过更新")
		helper := utils.NewResponseHelper(ctx)
		helper.Success("副本数未改变")
		return
	}

	originalReplicas := int32(0)
	if sts.Spec.Replicas != nil {
		originalReplicas = *sts.Spec.Replicas
	}

	sts.Spec.Replicas = &scaleReq.Replicas

	if sts.Annotations == nil {
		sts.Annotations = make(map[string]string)
	}
	sts.Annotations["scaled-by"] = "devops-console"
	sts.Annotations["scaled-at"] = metav1.Now().String()
	sts.Annotations["previous-replicas"] = fmt.Sprintf("%d", originalReplicas)

	_, err = client.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		if errors.IsConflict(err) {
			logs.Warning(logData, "更新冲突: "+err.Error())
			helper := utils.NewResponseHelper(ctx)
			helper.InternalError("资源版本冲突，请刷新后重试")
			return
		}
		logs.Error(map[string]interface{}{"replicas": scaleReq.Replicas, "error": err.Error(), "data": logData}, "扩缩容StatefulSet失败")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("扩缩容StatefulSet失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{"originalReplicas": originalReplicas, "newReplicas": scaleReq.Replicas, "data": logData}, "扩缩容StatefulSet成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("StatefulSet扩缩容成功", "data", map[string]interface{}{
		"name":      sts.Name,
		"namespace": sts.Namespace,
		"replicas":  scaleReq.Replicas,
	})
}

// DeleteStatefulSet 删除StatefulSet
func (c *StatefulSetController) DeleteStatefulSet(ctx *gin.Context) {
	namespace := ctx.Query("namespace")
	name := ctx.Query("name")

	logData := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	}
	logs.Debug(logData, "删除StatefulSet")

	instanceIDStr := ctx.Query("instanceId")
	instanceID := uint(1)
	if instanceIDStr != "" {
		if id, err := strconv.ParseInt(instanceIDStr, 10, 32); err == nil {
			instanceID = uint(id)
		}
	}

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		logs.Error(logData, "K8s客户端未初始化")
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	deletePolicy := metav1.DeletePropagationForeground
	err := client.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		logs.Error(logData, "删除StatefulSet失败: "+err.Error())
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除StatefulSet失败: " + err.Error())
		return
	}

	logs.Info(logData, "删除StatefulSet成功")
	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("StatefulSet删除成功", "data", map[string]interface{}{
		"name":      name,
		"namespace": namespace,
	})
}

// convertCreateRequestToK8sStatefulSet 转换创建请求为K8s StatefulSet
func (c *StatefulSetController) convertCreateRequestToK8sStatefulSet(namespace string, req k8s.StatefulSetCreateRequest) *appsv1.StatefulSet {
	if req.Labels == nil {
		req.Labels = map[string]string{
			"app": req.Name,
		}
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			Labels:    req.Labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &req.Replicas,
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

	if req.Port > 0 {
		sts.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: req.Port,
				Protocol:      corev1.ProtocolTCP,
			},
		}
	}

	return sts
}
