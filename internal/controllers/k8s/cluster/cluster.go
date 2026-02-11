package cluster

import (
	"devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// ClusterController Kubernetes集群控制器
type ClusterController struct{}

// NewClusterController 创建集群控制器实例
func NewClusterController() *ClusterController {
	return &ClusterController{}
}

// GetClusterList 获取集群列表
func (c *ClusterController) GetClusterList(ctx *gin.Context) {
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

	// 获取服务器版本
	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取服务器版本失败: " + err.Error())
		return
	}

	// 获取节点信息
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取节点信息失败: " + err.Error())
		return
	}

	// 统计节点状态
	totalNodes := len(nodes.Items)
	readyNodes := 0

	for _, node := range nodes.Items {
		// 检查节点状态
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				readyNodes++
				break
			}
		}
	}

	// 获取集群名称
	clusterName := "default-cluster"
	for _, node := range nodes.Items {
		if name, exists := node.Labels["kubernetes.io/cluster.name"]; exists {
			clusterName = name
			break
		}
	}

	// 构建集群列表响应
	clusterList := []map[string]interface{}{
		{
			"id":         instanceID,
			"name":       clusterName,
			"version":    serverVersion.GitVersion,
			"status":     "Running",
			"totalNodes": totalNodes,
			"readyNodes": readyNodes,
			"createTime": nodes.Items[0].CreationTimestamp.Format("2006-01-02 15:04:05"),
			"lastSync":   time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("获取集群列表成功", "data", clusterList)
}

// GetClusterInfo 获取集群基本信息
func (c *ClusterController) GetClusterInfo(ctx *gin.Context) {
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

	// 使用goroutine并行获取数据
	type result struct {
		serverVersion *version.Info
		nodes         *corev1.NodeList
		namespaces    *corev1.NamespaceList
		pods          *corev1.PodList
		err           error
	}

	ch := make(chan result, 4)

	// 并行获取服务器版本
	go func() {
		ver, err := client.Discovery().ServerVersion()
		ch <- result{serverVersion: ver, err: err}
	}()

	// 并行获取节点信息
	go func() {
		nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		ch <- result{nodes: nodes, err: err}
	}()

	// 并行获取命名空间信息
	go func() {
		ns, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		ch <- result{namespaces: ns, err: err}
	}()

	// 并行获取所有Pod（一次性获取，避免逐个命名空间查询）
	go func() {
		pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		ch <- result{pods: pods, err: err}
	}()

	// 收集结果
	var serverVersion *version.Info
	var nodes *corev1.NodeList
	var pods *corev1.PodList

	for i := 0; i < 4; i++ {
		r := <-ch
		if r.err != nil {
			continue
		}
		if r.serverVersion != nil {
			serverVersion = r.serverVersion
		}
		if r.nodes != nil {
			nodes = r.nodes
		}
		if r.pods != nil {
			pods = r.pods
		}
	}

	// 检查必要数据
	if serverVersion == nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取服务器版本失败")
		return
	}
	if nodes == nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取节点信息失败")
		return
	}

	// 统计节点状态
	totalNodes := len(nodes.Items)
	readyNodes := 0
	var masterNodes []string
	var workerNodes []string

	for _, node := range nodes.Items {
		// 检查节点状态
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				readyNodes++
				break
			}
		}

		// 判断节点角色
		if _, exists := node.Labels["node-role.kubernetes.io/master"]; exists {
			masterNodes = append(masterNodes, node.Name)
		} else if _, exists := node.Labels["node-role.kubernetes.io/control-plane"]; exists {
			masterNodes = append(masterNodes, node.Name)
		} else {
			workerNodes = append(workerNodes, node.Name)
		}
	}

	// 统计Pod信息（从已获取的pods列表中计算）
	podCount := 0
	runningPods := 0
	if pods != nil {
		podCount = len(pods.Items)
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				runningPods++
			}
		}
	}

	// 获取CPU和内存总量
	var totalCPU, totalMemory int64
	for _, node := range nodes.Items {
		if cpu := node.Status.Allocatable.Cpu(); !cpu.IsZero() {
			totalCPU += cpu.MilliValue()
		}
		if memory := node.Status.Allocatable.Memory(); !memory.IsZero() {
			totalMemory += memory.Value()
		}
	}

	// 获取网络配置
	serviceCIDR := "未配置"
	podCIDR := "未配置"

	// 尝试从kube-system命名空间的ConfigMap中获取网络配置
	if cm, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "cluster-config", metav1.GetOptions{}); err == nil {
		if cidr, exists := cm.Data["networking.serviceSubnet"]; exists {
			serviceCIDR = cidr
		}
		if cidr, exists := cm.Data["networking.podSubnet"]; exists {
			podCIDR = cidr
		}
	}

	// 获取API Server地址
	apiServer := "未配置"
	if client.Discovery().RESTClient() != nil {
		if url := client.Discovery().RESTClient().Get().URL(); url != nil && len(url.Host) > 0 {
			apiServer = "https://" + url.Host
		}
	}

	// 计算集群运行时间
	clusterAge := time.Since(nodes.Items[0].CreationTimestamp.Time)
	days := int(clusterAge.Hours() / 24)
	hours := int(clusterAge.Hours()) % 24
	uptime := fmt.Sprintf("%d天%d小时", days, hours)

	// 获取集群名称（从第一个节点的标签或配置中获取）
	clusterName := "default-cluster"
	for _, node := range nodes.Items {
		if name, exists := node.Labels["kubernetes.io/cluster.name"]; exists {
			clusterName = name
			break
		}
	}

	// 判断集群类型
	clusterType := "标准集群"
	if totalNodes <= 1 {
		clusterType = "单节点集群"
	} else if totalNodes <= 3 {
		clusterType = "小型集群"
	} else if totalNodes <= 10 {
		clusterType = "中型集群"
	} else {
		clusterType = "大型集群"
	}

	// 构建集群信息响应
	clusterInfo := k8s.ClusterInfo{
		Name:        clusterName,
		Version:     serverVersion.GitVersion,
		Type:        clusterType,
		CreateTime:  nodes.Items[0].CreationTimestamp.Format("2006-01-02 15:04:05"),
		LastSync:    time.Now().Format("2006-01-02 15:04:05"),
		Uptime:      uptime,
		TotalNodes:  totalNodes,
		ReadyNodes:  readyNodes,
		MasterNodes: masterNodes,
		WorkerNodes: workerNodes,
		TotalPods:   podCount,
		RunningPods: runningPods,
		CPUTotal:    float64(totalCPU) / 1000,    // 转换为核数
		MemoryTotal: totalMemory / (1024 * 1024), // 转换为Mi
	}

	// 获取网络插件信息
	networkPlugin := "未知"
	// 尝试从DaemonSet获取网络插件信息
	if daemonSets, err := client.AppsV1().DaemonSets("kube-system").List(ctx, metav1.ListOptions{}); err == nil {
		for _, ds := range daemonSets.Items {
			if ds.Name == "calico-node" {
				networkPlugin = "Calico"
				break
			} else if ds.Name == "flannel" {
				networkPlugin = "Flannel"
				break
			} else if ds.Name == "weave-net" {
				networkPlugin = "Weave"
				break
			}
		}
	}

	// 获取服务转发模式
	serviceForward := "未知"
	if cm, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "kube-proxy", metav1.GetOptions{}); err == nil {
		if mode, exists := cm.Data["mode"]; exists {
			serviceForward = mode
		}
	}

	// 获取DNS服务版本
	coreDnsVersion := "未知"
	if pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{LabelSelector: "k8s-app=kube-dns"}); err == nil && len(pods.Items) > 0 {
		for _, pod := range pods.Items {
			if len(pod.Spec.Containers) > 0 {
				for _, container := range pod.Spec.Containers {
					if container.Name == "coredns" {
						if parts := strings.Split(container.Image, ":"); len(parts) > 1 {
							coreDnsVersion = "v" + parts[1]
						}
						break
					}
				}
				break
			}
		}
	}

	// 获取etcd版本
	etcdVersion := "未知"
	if pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{LabelSelector: "component=etcd"}); err == nil && len(pods.Items) > 0 {
		for _, pod := range pods.Items {
			if len(pod.Spec.Containers) > 0 {
				for _, container := range pod.Spec.Containers {
					if container.Name == "etcd" {
						if parts := strings.Split(container.Image, ":"); len(parts) > 1 {
							etcdVersion = parts[1]
						}
						break
					}
				}
				break
			}
		}
	}

	// 获取kube-proxy版本
	kubeProxyVersion := "未知"
	if pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{LabelSelector: "component=kube-proxy"}); err == nil && len(pods.Items) > 0 {
		for _, pod := range pods.Items {
			if len(pod.Spec.Containers) > 0 {
				for _, container := range pod.Spec.Containers {
					if container.Name == "kube-proxy" {
						if parts := strings.Split(container.Image, ":"); len(parts) > 1 {
							kubeProxyVersion = "v" + parts[1]
						}
						break
					}
				}
				break
			}
		}
	}

	// 获取容器运行时
	containerRuntime := "未知"
	if len(nodes.Items) > 0 {
		containerRuntime = nodes.Items[0].Status.NodeInfo.ContainerRuntimeVersion
	}

	// 网络配置
	networkConfig := k8s.NetworkConfig{
		ServiceCidr:    serviceCIDR,
		PodCidr:        podCIDR,
		ApiServer:      apiServer,
		NetworkPlugin:  networkPlugin,
		ServiceForward: serviceForward,
		DnsService:     coreDnsVersion,
	}

	// 运行时信息
	runtimeInfo := k8s.RuntimeInfo{
		ContainerRuntime: containerRuntime,
		ApiServerVersion: serverVersion.GitVersion,
		EtcdVersion:      etcdVersion,
		CoreDnsVersion:   coreDnsVersion,
		KubeProxyVersion: kubeProxyVersion,
	}

	// 构建完整响应
	response := map[string]interface{}{
		"clusterInfo":   clusterInfo,
		"networkConfig": networkConfig,
		"runtimeInfo":   runtimeInfo,
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("获取集群信息成功", "data", response)
}

// GetClusterMetrics 获取集群指标数据
func (c *ClusterController) GetClusterMetrics(ctx *gin.Context) {
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

	// 使用goroutine并行获取数据
	type metricsResult struct {
		nodes   *corev1.NodeList
		pods    *corev1.PodList
		metrics *metricsv.NodeMetricsList
		err     error
	}

	ch := make(chan metricsResult, 3)

	// 并行获取节点信息
	go func() {
		nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		ch <- metricsResult{nodes: nodes, err: err}
	}()

	// 并行获取所有Pod信息
	go func() {
		pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		ch <- metricsResult{pods: pods, err: err}
	}()

	// 并行获取metrics数据（如果可用）
	go func() {
		if metricsClient, exists := configs.GetMetricsClient(instanceID); exists {
			metrics, err := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
			ch <- metricsResult{metrics: metrics, err: err}
		} else {
			ch <- metricsResult{err: fmt.Errorf("metrics client not available")}
		}
	}()

	// 收集结果
	var nodes *corev1.NodeList
	var pods *corev1.PodList
	var nodeMetrics *metricsv.NodeMetricsList

	for i := 0; i < 3; i++ {
		r := <-ch
		if r.err != nil {
			continue
		}
		if r.nodes != nil {
			nodes = r.nodes
		}
		if r.pods != nil {
			pods = r.pods
		}
		if r.metrics != nil {
			nodeMetrics = r.metrics
		}
	}

	// 检查必要数据
	if nodes == nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取节点信息失败")
		return
	}
	if pods == nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取Pod信息失败")
		return
	}

	// 统计资源使用情况
	var totalCPUAllocatable, totalMemoryAllocatable int64
	var totalCPUUsed, totalMemoryUsed int64
	readyNodes := 0

	// 创建节点metrics映射，便于快速查找
	nodeMetricsMap := make(map[string]metricsv.NodeMetrics)
	if nodeMetrics != nil {
		for _, metric := range nodeMetrics.Items {
			nodeMetricsMap[metric.Name] = metric
		}
	}

	for _, node := range nodes.Items {
		// 统计就绪节点
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				readyNodes++
				break
			}
		}

		// 统计资源总量
		if cpu := node.Status.Allocatable.Cpu(); !cpu.IsZero() {
			totalCPUAllocatable += cpu.MilliValue()
		}
		if memory := node.Status.Allocatable.Memory(); !memory.IsZero() {
			totalMemoryAllocatable += memory.Value()
		}

		// 优先使用metrics-server获取实际使用量
		if metric, exists := nodeMetricsMap[node.Name]; exists {
			// 使用实际使用量
			if !metric.Usage.Cpu().IsZero() {
				totalCPUUsed += metric.Usage.Cpu().MilliValue()
			}
			if !metric.Usage.Memory().IsZero() {
				totalMemoryUsed += metric.Usage.Memory().Value()
			}
		} else {
			// 如果没有metrics数据，使用Pod requests统计
			for _, pod := range pods.Items {
				if pod.Spec.NodeName == node.Name && pod.Status.Phase == corev1.PodRunning {
					for _, container := range pod.Spec.Containers {
						if cpu := container.Resources.Requests.Cpu(); !cpu.IsZero() {
							totalCPUUsed += cpu.MilliValue()
						}
						if memory := container.Resources.Requests.Memory(); !memory.IsZero() {
							totalMemoryUsed += memory.Value()
						}
					}
				}
			}
		}
	}

	// 计算使用率
	cpuUsage := float64(0)
	memoryUsage := float64(0)
	if totalCPUAllocatable > 0 {
		cpuUsage = float64(totalCPUUsed) / float64(totalCPUAllocatable) * 100
	}
	if totalMemoryAllocatable > 0 {
		memoryUsage = float64(totalMemoryUsed) / float64(totalMemoryAllocatable) * 100
	}

	// 获取工作负载统计
	workloadStats := c.getWorkloadStats(ctx, client)

	// 获取存储信息
	storageInfo := c.getStorageInfo(ctx, client)

	// 构建指标响应
	metrics := k8s.ClusterMetrics{
		TotalNodes:      len(nodes.Items),
		ReadyNodes:      readyNodes,
		TotalPods:       len(pods.Items),
		CpuUsage:        int(cpuUsage),
		CpuAvailable:    float64(totalCPUAllocatable-totalCPUUsed) / 1000,
		CpuTotal:        float64(totalCPUAllocatable) / 1000,
		MemoryUsage:     int(memoryUsage),
		MemoryAvailable: (totalMemoryAllocatable - totalMemoryUsed) / (1024 * 1024),
		MemoryTotal:     totalMemoryAllocatable / (1024 * 1024),
		WorkloadStats:   workloadStats,
		StorageInfo:     storageInfo,
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("获取集群指标成功", "metrics", metrics)
}

// GetNodeList 获取节点列表
func (c *ClusterController) GetNodeList(ctx *gin.Context) {
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

	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		helper := utils.NewResponseHelper(ctx)
		helper.InternalError("获取节点列表失败: " + err.Error())
		return
	}

	var nodeList []k8s.NodeInfo
	for _, node := range nodes.Items {
		// 获取节点内部IP
		internalIP := ""
		externalIP := ""
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				internalIP = addr.Address
			} else if addr.Type == corev1.NodeExternalIP {
				externalIP = addr.Address
			}
		}

		// 获取节点状态
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

		// 判断节点角色
		role := "worker"
		if _, exists := node.Labels["node-role.kubernetes.io/master"]; exists {
			role = "master"
		} else if _, exists := node.Labels["node-role.kubernetes.io/control-plane"]; exists {
			role = "master"
		}

		// 获取Pod数量
		pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
		})
		podCount := 0
		if err == nil {
			podCount = len(pods.Items)
		}

		// 获取资源使用率（通过metrics-server API获取）
		cpuUsage := 0
		memoryUsage := 0

		// 尝试获取节点指标
		if metricsClient, exists := configs.GetMetricsClient(instanceID); exists {
			if nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, metav1.GetOptions{}); err == nil {
				// 计算CPU使用率
				if cpuAlloc := node.Status.Allocatable.Cpu(); !cpuAlloc.IsZero() && !nodeMetrics.Usage.Cpu().IsZero() {
					cpuUsed := nodeMetrics.Usage.Cpu().MilliValue()
					cpuTotal := cpuAlloc.MilliValue()
					if cpuTotal > 0 {
						cpuUsage = int(float64(cpuUsed) / float64(cpuTotal) * 100)
					}
				}

				// 计算内存使用率
				if memAlloc := node.Status.Allocatable.Memory(); !memAlloc.IsZero() && !nodeMetrics.Usage.Memory().IsZero() {
					memUsed := nodeMetrics.Usage.Memory().Value()
					memTotal := memAlloc.Value()
					if memTotal > 0 {
						memoryUsage = int(float64(memUsed) / float64(memTotal) * 100)
					}
				}
			}
		}

		nodeInfo := k8s.NodeInfo{
			Name:               node.Name,
			Status:             status,
			Role:               role,
			InternalIP:         internalIP,
			ExternalIP:         externalIP,
			CpuUsage:           cpuUsage,
			MemoryUsage:        memoryUsage,
			PodCount:           podCount,
			PodCapacity:        node.Status.Allocatable.Pods().String(),
			K8sVersion:         node.Status.NodeInfo.KubeletVersion,
			CreateTime:         node.CreationTimestamp.Unix(),
			OsImage:            node.Status.NodeInfo.OSImage,
			KernelVersion:      node.Status.NodeInfo.KernelVersion,
			ContainerRuntime:   node.Status.NodeInfo.ContainerRuntimeVersion,
			KubeletVersion:     node.Status.NodeInfo.KubeletVersion,
			CpuCapacity:        node.Status.Capacity.Cpu().String(),
			CpuAllocatable:     node.Status.Allocatable.Cpu().String(),
			MemoryCapacity:     node.Status.Capacity.Memory().String(),
			MemoryAllocatable:  node.Status.Allocatable.Memory().String(),
			StorageCapacity:    node.Status.Capacity.StorageEphemeral().String(),
			StorageAllocatable: node.Status.Allocatable.StorageEphemeral().String(),
			Labels:             node.Labels,
			Annotations:        node.Annotations,
		}

		nodeList = append(nodeList, nodeInfo)
	}

	helper := utils.NewResponseHelper(ctx)
	helper.SuccessWithData("获取节点列表成功", "nodeList", nodeList)
}

// getWorkloadStats 获取工作负载统计
func (c *ClusterController) getWorkloadStats(ctx *gin.Context, client *kubernetes.Clientset) k8s.WorkloadStats {
	stats := k8s.WorkloadStats{}

	// 获取Deployments
	if deployments, err := client.AppsV1().Deployments("").List(ctx, metav1.ListOptions{}); err == nil {
		stats.Deployments = len(deployments.Items)
	}

	// 获取StatefulSets
	if statefulSets, err := client.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{}); err == nil {
		stats.StatefulSets = len(statefulSets.Items)
	}

	// 获取DaemonSets
	if daemonSets, err := client.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{}); err == nil {
		stats.DaemonSets = len(daemonSets.Items)
	}

	// 获取Jobs
	if jobs, err := client.BatchV1().Jobs("").List(ctx, metav1.ListOptions{}); err == nil {
		stats.Jobs = len(jobs.Items)
	}

	// 获取Pods
	if pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{}); err == nil {
		stats.TotalPods = len(pods.Items)
		for _, pod := range pods.Items {
			switch pod.Status.Phase {
			case corev1.PodRunning:
				stats.RunningPods++
			case corev1.PodPending:
				stats.PendingPods++
			case corev1.PodFailed:
				stats.FailedPods++
			case corev1.PodSucceeded:
				stats.SucceededPods++
			case corev1.PodUnknown:
				stats.UnknownPods++
			}
		}
	}

	return stats
}

// getStorageInfo 获取存储信息
func (c *ClusterController) getStorageInfo(ctx *gin.Context, client *kubernetes.Clientset) k8s.StorageInfo {
	info := k8s.StorageInfo{}

	// 获取PV
	if pvs, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{}); err == nil {
		info.TotalPV = len(pvs.Items)
	}

	// 获取PVC
	if pvcs, err := client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{}); err == nil {
		info.TotalPVC = len(pvcs.Items)
	}

	// 获取StorageClasses
	if storageClasses, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{}); err == nil {
		info.StorageClasses = len(storageClasses.Items)
	}

	// 计算已使用存储
	usedStorage := int64(0)
	if pvcs, err := client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{}); err == nil {
		for _, pvc := range pvcs.Items {
			if pvc.Status.Capacity != nil {
				if storage := pvc.Status.Capacity.Storage(); !storage.IsZero() {
					usedStorage += storage.Value() / (1024 * 1024 * 1024) // 转换为GB
				}
			}
		}
	}
	info.UsedStorage = int(usedStorage)

	return info
}
