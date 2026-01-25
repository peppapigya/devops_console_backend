package pod

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodController Pod控制器
type PodController struct{}

// NewPodController 创建Pod控制器实例
func NewPodController() *PodController {
	return &PodController{}
}

// GetPodDetail 获取Pod详情
func (c *PodController) GetPodDetail(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	podName := ctx.Param("podname")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
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

	podDetail, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		// 尝试在所有命名空间中查找该Pod
		allPods, err2 := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", podName),
		})

		if err2 == nil && len(allPods.Items) > 0 {
			// 找到了Pod，返回实际命名空间的错误信息
			actualNamespace := allPods.Items[0].Namespace
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest(fmt.Sprintf("Pod '%s' 存在于命名空间 '%s' 中，而不是 '%s'", podName, actualNamespace, namespace))
			return
		}

		helper := utils.NewResponseHelper(ctx)
		helper.NotFound(fmt.Sprintf("Pod '%s' 在命名空间 '%s' 中不存在", podName, namespace))
		return
	}

	// 转换容器信息
	containers := make([]gin.H, 0)
	for _, container := range podDetail.Spec.Containers {
		containerInfo := gin.H{
			"name":            container.Name,
			"image":           container.Image,
			"imagePullPolicy": string(container.ImagePullPolicy),
			"ports":           container.Ports,
			"env":             container.Env,
			"resources":       container.Resources,
		}
		containers = append(containers, containerInfo)
	}

	// 计算重启次数
	restartCount := 0
	for _, containerStatus := range podDetail.Status.ContainerStatuses {
		restartCount += int(containerStatus.RestartCount)
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "podDetail", gin.H{
		"name":          podDetail.Name,
		"namespace":     podDetail.Namespace,
		"uid":           string(podDetail.UID),
		"status":        podDetail.Status.Phase,
		"restartPolicy": string(podDetail.Spec.RestartPolicy),
		"restarts":      restartCount,
		"age":           podDetail.CreationTimestamp.Unix(),
		"ip":            podDetail.Status.PodIP,
		"node":          podDetail.Spec.NodeName,
		"labels":        podDetail.Labels,
		"annotations":   podDetail.Annotations,
		"containers":    containers,
		"conditions":    podDetail.Status.Conditions,
	})
}

// GetPodList 获取Pod列表
func (c *PodController) GetPodList(ctx *gin.Context) {
	namespace := ctx.Param("namespace")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
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

	var list *corev1.PodList
	var err error

	// 如果是all，获取所有命名空间的Pod
	if namespace == "all" {
		list, err = client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Pod列表失败")
		return
	}

	podList := make([]k8s.PodListItem, 0)
	for _, item := range list.Items {
		podItem := c.convertPodToListItem(item)
		podList = append(podList, podItem)
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "podList", podList)
}

// CreatePod 创建Pod
func (c *PodController) CreatePod(ctx *gin.Context) {
	var req k8s.PodCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
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

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	var pod *corev1.Pod
	var err error

	// 如果提供了YAML，直接使用YAML创建
	if req.YAML != "" {
		pod, err = c.parseYAMLToPod(req.YAML)
		if err != nil {
			helper := utils.NewResponseHelper(ctx)
			helper.BadRequest("YAML解析失败: " + err.Error())
			return
		}
	} else {
		// 使用表单数据创建
		pod = c.convertCreateRequestToK8sPod(&req)
	}

	_, err = client.CoreV1().Pods(req.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("创建Pod失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Pod创建成功")
}

// UpdatePod 更新Pod
func (c *PodController) UpdatePod(ctx *gin.Context) {
	var req k8s.PodUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
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

	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 获取现有Pod
	podNS := client.CoreV1().Pods(req.Namespace)
	newPodDetail, err := podNS.Get(ctx, req.Podname, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound("Pod不存在")
		return
	}

	// 更新注解
	if newPodDetail.Annotations == nil {
		newPodDetail.Annotations = make(map[string]string)
	}
	newPodDetail.Annotations["updated-by"] = "update pod"

	// 更新容器信息
	containerFound := false
	for i, container := range newPodDetail.Spec.Containers {
		if container.Name == req.Imagename {
			newPodDetail.Spec.Containers[i].Image = req.Image
			containerFound = true
			break
		}
	}

	if !containerFound {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("容器 " + req.Imagename + " 不存在")
		return
	}

	_, err = podNS.Update(ctx, newPodDetail, metav1.UpdateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("更新Pod失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Pod更新成功")
}

// GetPodEvents 获取Pod事件
func (c *PodController) GetPodEvents(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	podName := ctx.Param("podname")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
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

	// 获取Pod事件
	fieldSelector := fmt.Sprintf("involvedObject.name=%s", podName)
	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
		Limit:         50,
	})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Pod事件失败: " + err.Error())
		return
	}

	// 转换事件数据
	eventList := make([]gin.H, 0)
	for _, event := range events.Items {
		eventData := gin.H{
			"type":           event.Type,
			"reason":         event.Reason,
			"message":        event.Message,
			"source":         event.Source,
			"firstTimestamp": event.FirstTimestamp,
			"lastTimestamp":  event.LastTimestamp,
			"count":          event.Count,
		}
		eventList = append(eventList, eventData)
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "events", eventList)
}

// GetPodLogs 获取Pod日志
func (c *PodController) GetPodLogs(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	podName := ctx.Param("podname")
	container := ctx.Query("container")
	follow := ctx.Query("follow") == "true"
	tailLinesStr := ctx.Query("tail_lines")
	timestamps := ctx.Query("timestamps") == "true"

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
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

	tailLines := int64(100)
	if tailLinesStr != "" {
		if lines, err := strconv.ParseInt(tailLinesStr, 10, 64); err == nil {
			tailLines = lines
		}
	}

	// 获取Pod日志请求
	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container:  container,
		Follow:     follow,
		TailLines:  &tailLines,
		Timestamps: timestamps,
		Previous:   false,
	})

	logs, err := req.Stream(ctx)
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Pod日志失败: " + err.Error())
		return
	}
	defer logs.Close()

	// 读取日志内容
	var logContent strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := logs.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			helper := utils.NewResponseHelper(ctx)
			helper.InternalError("读取日志失败: " + err.Error())
			return
		}
		if n > 0 {
			logContent.Write(buf[:n])
		}
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "logs", gin.H{
		"logs": logContent.String(),
	})
}

// DeletePod 删除Pod
func (c *PodController) DeletePod(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	podName := ctx.Param("podname")

	instanceIDStr := ctx.Query("instance_id")
	instanceID := uint(1) // 默认值
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

	err := client.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除Pod失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("Pod删除成功")
}

// convertPodToListItem 转换K8s Pod为列表项
func (c *PodController) convertPodToListItem(pod corev1.Pod) k8s.PodListItem {
	var totalC, readyC, restartC int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Ready {
			readyC++
		}
		restartC += containerStatus.RestartCount
		totalC++
	}

	var podStatus string
	if pod.Status.Phase != "Running" {
		podStatus = "Error"
	} else {
		podStatus = "Running"
	}

	return k8s.PodListItem{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Ready:     fmt.Sprintf("%d/%d", readyC, totalC),
		Status:    podStatus,
		Restarts:  restartC,
		Age:       pod.CreationTimestamp.Unix(),
		IP:        pod.Status.PodIP,
		Node:      pod.Spec.NodeName,
	}
}

// getRestartCount 获取Pod总重启次数
func (c *PodController) getRestartCount(pod *corev1.Pod) int32 {
	var restartCount int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		restartCount += containerStatus.RestartCount
	}
	return restartCount
}

// parseYAMLToPod 解析YAML为Pod对象
func (c *PodController) parseYAMLToPod(yamlContent string) (*corev1.Pod, error) {
	var pod corev1.Pod
	err := yaml.Unmarshal([]byte(yamlContent), &pod)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

// convertCreateRequestToK8sPod 转换创建请求为K8s Pod
func (c *PodController) convertCreateRequestToK8sPod(req *k8s.PodCreateRequest) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Podname,
			Namespace: req.Namespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicy(req.RestartPolicy),
		},
	}

	// 添加标签
	if len(req.Labels) > 0 {
		labels := make(map[string]string)
		for _, label := range req.Labels {
			if label.Key != "" && label.Value != "" {
				labels[label.Key] = label.Value
			}
		}
		pod.ObjectMeta.Labels = labels
	}

	// 添加注解
	if len(req.Annotations) > 0 {
		annotations := make(map[string]string)
		for _, annotation := range req.Annotations {
			if annotation.Key != "" && annotation.Value != "" {
				annotations[annotation.Key] = annotation.Value
			}
		}
		pod.ObjectMeta.Annotations = annotations
	}

	// 添加节点选择器
	if len(req.NodeSelector) > 0 {
		nodeSelector := make(map[string]string)
		for _, selector := range req.NodeSelector {
			if selector.Key != "" && selector.Value != "" {
				nodeSelector[selector.Key] = selector.Value
			}
		}
		pod.Spec.NodeSelector = nodeSelector
	}

	// 转换容器配置
	containers := make([]corev1.Container, 0, len(req.Containers))
	for _, container := range req.Containers {
		k8sContainer := corev1.Container{
			Name:            container.Name,
			Image:           container.Image,
			ImagePullPolicy: corev1.PullPolicy(container.ImagePullPolicy),
		}

		// 添加端口
		if len(container.Ports) > 0 {
			ports := make([]corev1.ContainerPort, 0, len(container.Ports))
			for _, port := range container.Ports {
				if port.ContainerPort > 0 {
					ports = append(ports, corev1.ContainerPort{
						Name:          port.Name,
						ContainerPort: port.ContainerPort,
						Protocol:      corev1.Protocol(strings.ToUpper(port.Protocol)),
					})
				}
			}
			k8sContainer.Ports = ports
		}

		// 添加环境变量
		if len(container.Env) > 0 {
			env := make([]corev1.EnvVar, 0, len(container.Env))
			for _, envVar := range container.Env {
				if envVar.Name != "" {
					env = append(env, corev1.EnvVar{
						Name:  envVar.Name,
						Value: envVar.Value,
					})
				}
			}
			k8sContainer.Env = env
		}

		// 添加资源限制
		resources := corev1.ResourceRequirements{}
		requests := make(corev1.ResourceList)
		limits := make(corev1.ResourceList)

		if container.Resources.Requests.CPU != "" {
			if cpu, err := resource.ParseQuantity(container.Resources.Requests.CPU); err == nil {
				requests[corev1.ResourceCPU] = cpu
			}
		}
		if container.Resources.Requests.Memory != "" {
			if memory, err := resource.ParseQuantity(container.Resources.Requests.Memory); err == nil {
				requests[corev1.ResourceMemory] = memory
			}
		}
		if container.Resources.Limits.CPU != "" {
			if cpu, err := resource.ParseQuantity(container.Resources.Limits.CPU); err == nil {
				limits[corev1.ResourceCPU] = cpu
			}
		}
		if container.Resources.Limits.Memory != "" {
			if memory, err := resource.ParseQuantity(container.Resources.Limits.Memory); err == nil {
				limits[corev1.ResourceMemory] = memory
			}
		}

		if len(requests) > 0 || len(limits) > 0 {
			resources.Requests = requests
			resources.Limits = limits
			k8sContainer.Resources = resources
		}

		containers = append(containers, k8sContainer)
	}

	pod.Spec.Containers = containers
	return pod
}
