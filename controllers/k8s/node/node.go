package node

import (
	"context"
	"devops-console-backend/config"
	"devops-console-backend/models/k8s"
	"devops-console-backend/pkg/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NodeController Node控制器
type NodeController struct{}

// NewNodeController 创建Node控制器实例
func NewNodeController() *NodeController {
	return &NodeController{}
}

// GetNodeList 获取节点列表
func (c *NodeController) GetNodeList(ctx *gin.Context) {
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

	nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取节点列表失败: " + err.Error())
		return
	}

	nodes := make([]k8s.NodeListItem, 0)
	for _, node := range nodeList.Items {
		// 获取节点上的Pod数量
		podList, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
		})
		podCount := 0
		if err == nil {
			podCount = len(podList.Items)
		}

		nodeItem := c.convertNodeToListItemWithPodCount(node, podCount)
		nodes = append(nodes, nodeItem)
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "nodeList", nodes)
}

// GetNodeDetail 获取节点详情
func (c *NodeController) GetNodeDetail(ctx *gin.Context) {
	nodeName := ctx.Param("nodeName")

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

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound(fmt.Sprintf("节点 '%s' 不存在", nodeName))
		return
	}

	// 获取节点上的Pod
	podList, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取节点Pod列表失败: " + err.Error())
		return
	}

	pods := make([]gin.H, 0)
	for _, pod := range podList.Items {
		podInfo := gin.H{
			"name":      pod.Name,
			"namespace": pod.Namespace,
			"status":    pod.Status.Phase,
			"created":   pod.CreationTimestamp.Unix(),
		}
		pods = append(pods, podInfo)
	}

	// 获取PodCIDR信息
	podCIDR := ""
	if node.Spec.PodCIDR != "" {
		podCIDR = node.Spec.PodCIDR
	} else if node.Labels != nil {
		podCIDR = node.Labels["kubernetes.io/pod-cidr"]
	}

	// 检查网络策略配置
	networkPolicyAvailable := false
	if node.Labels != nil {
		for key := range node.Labels {
			if strings.Contains(key, "networking.k8s.io/policy") ||
				strings.Contains(key, "policy.beta.kubernetes.io") {
				networkPolicyAvailable = true
				break
			}
		}
	}

	// 获取K8s版本信息
	k8sVersion := node.Status.NodeInfo.KubeletVersion
	if k8sVersion == "" {
		k8sVersion = node.Status.NodeInfo.KubeProxyVersion
	}

	nodeDetail := gin.H{
		"name":              node.Name,
		"uid":               string(node.UID),
		"creationTimestamp": node.CreationTimestamp.Unix(),
		"labels":            node.Labels,
		"annotations":       node.Annotations,
		"status":            c.getNodeStatus(node),
		"conditions":        c.getNodeConditions(node.Status.Conditions),
		"addresses":         node.Status.Addresses,
		"nodeInfo":          node.Status.NodeInfo,
		"capacity":          node.Status.Capacity,
		"allocatable":       node.Status.Allocatable,
		"pods":              pods,
		"cordoned":          node.Spec.Unschedulable,
		// 添加资源信息
		"cpuCapacity":              node.Status.Capacity.Cpu().String(),
		"cpuAllocatable":           node.Status.Allocatable.Cpu().String(),
		"memoryCapacity":           node.Status.Capacity.Memory().String(),
		"memoryAllocatable":        node.Status.Allocatable.Memory().String(),
		"storageCapacity":          node.Status.Capacity.StorageEphemeral().String(),
		"storageAllocatable":       node.Status.Allocatable.StorageEphemeral().String(),
		"ephemeralStorageCapacity": node.Status.Capacity.StorageEphemeral().String(),
		"podCapacity":              node.Status.Capacity.Pods().String(),
		"podCount":                 len(pods),
		// 添加系统信息
		"osImage":          node.Status.NodeInfo.OSImage,
		"kernelVersion":    node.Status.NodeInfo.KernelVersion,
		"containerRuntime": node.Status.NodeInfo.ContainerRuntimeVersion,
		"kubeletVersion":   node.Status.NodeInfo.KubeletVersion,
		"kubeProxyVersion": node.Status.NodeInfo.KubeProxyVersion,
		"k8sVersion":       k8sVersion,
		"systemUUID":       node.Status.NodeInfo.SystemUUID,
		"machineID":        node.Status.NodeInfo.MachineID,
		"bootID":           node.Status.NodeInfo.BootID,
		"operatingSystem":  node.Status.NodeInfo.OperatingSystem,
		"architecture":     node.Status.NodeInfo.Architecture,
		"createTime":       node.CreationTimestamp.Unix(),
		// 添加网络信息
		"podCIDR":                podCIDR,
		"networkPolicyAvailable": networkPolicyAvailable,
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("success", "nodeDetail", nodeDetail)
}

// CordonNode 隔离节点
func (c *NodeController) CordonNode(ctx *gin.Context) {
	nodeName := ctx.Param("nodeName")

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

	// 获取当前节点状态
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound(fmt.Sprintf("节点 '%s' 不存在", nodeName))
		return
	}

	// 设置节点为不可调度
	node.Spec.Unschedulable = true

	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("隔离节点失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("节点隔离成功")
}

// UncordonNode 取消隔离节点
func (c *NodeController) UncordonNode(ctx *gin.Context) {
	nodeName := ctx.Param("nodeName")

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

	// 获取当前节点状态
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound(fmt.Sprintf("节点 '%s' 不存在", nodeName))
		return
	}

	// 设置节点为可调度
	node.Spec.Unschedulable = false

	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("取消隔离节点失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("取消隔离节点成功")
}

// DrainNode 排空节点
func (c *NodeController) DrainNode(ctx *gin.Context) {
	nodeName := ctx.Param("nodeName")

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

	// 首先隔离节点
	cordonErr := c.cordonNodeInternal(client, nodeName)
	if cordonErr != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("隔离节点失败: " + cordonErr.Error())
		return
	}

	// 获取节点上的Pod
	podList, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取节点Pod列表失败: " + err.Error())
		return
	}

	// 驱逐节点上的Pod
	for _, pod := range podList.Items {
		// 跳过系统Pod和静态Pod
		if pod.Namespace == "kube-system" || pod.Annotations["kubernetes.io/config.source"] == "file" {
			continue
		}

		// 执行驱逐操作
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
		}
		err := client.PolicyV1().Evictions(pod.Namespace).Evict(ctx, eviction)
		if err != nil {
			// 记录错误但继续处理其他Pod
			fmt.Printf("驱逐Pod %s/%s 失败: %v\n", pod.Namespace, pod.Name, err)
		}
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("节点排空操作已启动")
}

// AddNodeLabel 添加节点标签
func (c *NodeController) AddNodeLabel(ctx *gin.Context) {
	nodeName := ctx.Param("nodeName")

	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

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

	client, exists := config.GetK8sClient(instanceID)
	if !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("K8s客户端未初始化")
		return
	}

	// 获取当前节点
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound(fmt.Sprintf("节点 '%s' 不存在", nodeName))
		return
	}

	// 添加标签
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	node.Labels[req.Key] = req.Value

	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("添加节点标签失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("节点标签添加成功")
}

// RemoveNodeLabel 删除节点标签
func (c *NodeController) RemoveNodeLabel(ctx *gin.Context) {
	nodeName := ctx.Param("nodeName")
	labelKey := ctx.Param("labelKey")

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

	// 获取当前节点
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.NotFound(fmt.Sprintf("节点 '%s' 不存在", nodeName))
		return
	}

	// 删除标签
	if node.Labels == nil {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest("节点没有标签")
		return
	}

	if _, exists := node.Labels[labelKey]; !exists {
		helper := utils.NewResponseHelper(ctx)
		helper.BadRequest(fmt.Sprintf("标签 '%s' 不存在", labelKey))
		return
	}

	delete(node.Labels, labelKey)

	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("删除节点标签失败: " + err.Error())
		return
	}

	helper := utils.NewResponseHelper(ctx)
	helper.Success("节点标签删除成功")
}

// convertNodeToListItem 转换K8s Node为列表项
func (c *NodeController) convertNodeToListItem(node corev1.Node) k8s.NodeListItem {
	status := "Unknown"
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
			break
		}
	}

	var roles []string
	for k, v := range node.Labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") && v == "true" {
			role := strings.TrimPrefix(k, "node-role.kubernetes.io/")
			if role == "master" {
				role = "control-plane"
			}
			roles = append(roles, role)
		}
	}

	if len(roles) == 0 {
		roles = append(roles, "worker")
	}

	var internalIP, externalIP string
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			internalIP = addr.Address
		}
		if addr.Type == corev1.NodeExternalIP {
			externalIP = addr.Address
		}
	}

	var version string
	if node.Status.NodeInfo.KubeletVersion != "" {
		version = node.Status.NodeInfo.KubeletVersion
	} else if node.Status.NodeInfo.KubeProxyVersion != "" {
		version = node.Status.NodeInfo.KubeProxyVersion
	}

	// 获取Pod数量
	podCount := 0
	podCapacity := "110" // 默认值
	if node.Status.Capacity.Pods() != nil {
		podCapacity = node.Status.Capacity.Pods().String()
	}

	return k8s.NodeListItem{
		Name:        node.Name,
		Status:      status,
		Roles:       strings.Join(roles, ","),
		InternalIP:  internalIP,
		ExternalIP:  externalIP,
		Version:     version,
		Age:         node.CreationTimestamp.Unix(),
		Labels:      node.Labels,
		Annotations: node.Annotations,
		Cordoned:    node.Spec.Unschedulable,
		// 添加资源信息
		CPUCapacity:       node.Status.Capacity.Cpu().String(),
		CPUAllocatable:    node.Status.Allocatable.Cpu().String(),
		MemoryCapacity:    node.Status.Capacity.Memory().String(),
		MemoryAllocatable: node.Status.Allocatable.Memory().String(),
		StorageCapacity:   node.Status.Capacity.StorageEphemeral().String(),
		PodCapacity:       podCapacity,
		PodCount:          podCount,
		// 添加系统信息
		OSImage:          node.Status.NodeInfo.OSImage,
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
		KubeProxyVersion: node.Status.NodeInfo.KubeProxyVersion,
		SystemUUID:       node.Status.NodeInfo.SystemUUID,
		CreateTime:       node.CreationTimestamp.Unix(),
	}
}

// getNodeStatus 获取节点状态
func (c *NodeController) getNodeStatus(node *corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return string(condition.Reason)
		}
	}
	return "Unknown"
}

// getNodeConditions 获取节点条件
func (c *NodeController) getNodeConditions(conditions []corev1.NodeCondition) []gin.H {
	result := make([]gin.H, 0)
	for _, condition := range conditions {
		conditionInfo := gin.H{
			"type":           string(condition.Type),
			"status":         string(condition.Status),
			"reason":         condition.Reason,
			"message":        condition.Message,
			"lastHeartbeat":  condition.LastHeartbeatTime.Unix(),
			"lastTransition": condition.LastTransitionTime.Unix(),
		}
		result = append(result, conditionInfo)
	}
	return result
}

// convertNodeToListItemWithPodCount 转换K8s Node为列表项（带Pod数量）
func (c *NodeController) convertNodeToListItemWithPodCount(node corev1.Node, podCount int) k8s.NodeListItem {
	status := "Unknown"
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
			break
		}
	}

	var roles []string
	for k, v := range node.Labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") && v == "true" {
			role := strings.TrimPrefix(k, "node-role.kubernetes.io/")
			if role == "master" {
				role = "control-plane"
			}
			roles = append(roles, role)
		}
	}

	if len(roles) == 0 {
		roles = append(roles, "worker")
	}

	var internalIP, externalIP string
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			internalIP = addr.Address
		}
		if addr.Type == corev1.NodeExternalIP {
			externalIP = addr.Address
		}
	}

	var version string
	if node.Status.NodeInfo.KubeletVersion != "" {
		version = node.Status.NodeInfo.KubeletVersion
	} else if node.Status.NodeInfo.KubeProxyVersion != "" {
		version = node.Status.NodeInfo.KubeProxyVersion
	}

	// 获取Pod容量
	podCapacity := "110" // 默认值
	if node.Status.Capacity.Pods() != nil {
		podCapacity = node.Status.Capacity.Pods().String()
	}

	return k8s.NodeListItem{
		Name:        node.Name,
		Status:      status,
		Roles:       strings.Join(roles, ","),
		InternalIP:  internalIP,
		ExternalIP:  externalIP,
		Version:     version,
		Age:         node.CreationTimestamp.Unix(),
		Labels:      node.Labels,
		Annotations: node.Annotations,
		Cordoned:    node.Spec.Unschedulable,
		// 添加资源信息
		CPUCapacity:       node.Status.Capacity.Cpu().String(),
		CPUAllocatable:    node.Status.Allocatable.Cpu().String(),
		MemoryCapacity:    node.Status.Capacity.Memory().String(),
		MemoryAllocatable: node.Status.Allocatable.Memory().String(),
		StorageCapacity:   node.Status.Capacity.StorageEphemeral().String(),
		PodCapacity:       podCapacity,
		PodCount:          podCount,
		// 添加系统信息
		OSImage:          node.Status.NodeInfo.OSImage,
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
		KubeProxyVersion: node.Status.NodeInfo.KubeProxyVersion,
		SystemUUID:       node.Status.NodeInfo.SystemUUID,
		CreateTime:       node.CreationTimestamp.Unix(),
	}
}

// cordonNodeInternal 内部隔离节点方法
func (c *NodeController) cordonNodeInternal(client kubernetes.Interface, nodeName string) error {
	node, err := client.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	node.Spec.Unschedulable = true

	_, err = client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	return err
}
