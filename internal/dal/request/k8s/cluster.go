package k8s

// ClusterInfo 集群基本信息
type ClusterInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Type        string   `json:"type"`
	CreateTime  string   `json:"createTime"`
	LastSync    string   `json:"lastSync"`
	Uptime      string   `json:"uptime"`
	TotalNodes  int      `json:"totalNodes"`
	ReadyNodes  int      `json:"readyNodes"`
	MasterNodes []string `json:"masterNodes"`
	WorkerNodes []string `json:"workerNodes"`
	TotalPods   int      `json:"totalPods"`
	RunningPods int      `json:"runningPods"`
	CPUTotal    float64  `json:"cpuTotal"`
	MemoryTotal int64    `json:"memoryTotal"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	ServiceCidr    string `json:"serviceCidr"`
	PodCidr        string `json:"podCidr"`
	ApiServer      string `json:"apiServer"`
	NetworkPlugin  string `json:"networkPlugin"`
	ServiceForward string `json:"serviceForward"`
	DnsService     string `json:"dnsService"`
}

// RuntimeInfo 运行时信息
type RuntimeInfo struct {
	ContainerRuntime string `json:"containerRuntime"`
	ApiServerVersion string `json:"apiServerVersion"`
	EtcdVersion      string `json:"etcdVersion"`
	CoreDnsVersion   string `json:"coreDnsVersion"`
	KubeProxyVersion string `json:"kubeProxyVersion"`
}

// ClusterMetrics 集群指标
type ClusterMetrics struct {
	TotalNodes      int           `json:"totalNodes"`
	ReadyNodes      int           `json:"readyNodes"`
	TotalPods       int           `json:"totalPods"`
	CpuUsage        int           `json:"cpuUsage"`
	CpuAvailable    float64       `json:"cpuAvailable"`
	CpuTotal        float64       `json:"cpuTotal"`
	MemoryUsage     int           `json:"memoryUsage"`
	MemoryAvailable int64         `json:"memoryAvailable"`
	MemoryTotal     int64         `json:"memoryTotal"`
	WorkloadStats   WorkloadStats `json:"workloadStats"`
	StorageInfo     StorageInfo   `json:"storageInfo"`
}

// WorkloadStats 工作负载统计
type WorkloadStats struct {
	Deployments   int `json:"deployments"`
	StatefulSets  int `json:"statefulSets"`
	DaemonSets    int `json:"daemonSets"`
	Jobs          int `json:"jobs"`
	TotalPods     int `json:"totalPods"`
	RunningPods   int `json:"runningPods"`
	SucceededPods int `json:"succeededPods"`
	FailedPods    int `json:"failedPods"`
	PendingPods   int `json:"pendingPods"`
	UnknownPods   int `json:"unknownPods"`
}

// StorageInfo 存储信息
type StorageInfo struct {
	TotalPV        int `json:"totalPV"`
	TotalPVC       int `json:"totalPVC"`
	StorageClasses int `json:"storageClasses"`
	UsedStorage    int `json:"usedStorage"`
}

// NodeInfo 节点信息
type NodeInfo struct {
	Name               string            `json:"name"`
	Status             string            `json:"status"`
	Role               string            `json:"role"`
	InternalIP         string            `json:"internalIP"`
	ExternalIP         string            `json:"externalIP"`
	CpuUsage           int               `json:"cpuUsage"`
	MemoryUsage        int               `json:"memoryUsage"`
	PodCount           int               `json:"podCount"`
	PodCapacity        string            `json:"podCapacity"`
	K8sVersion         string            `json:"k8sVersion"`
	CreateTime         int64             `json:"createTime"`
	OsImage            string            `json:"osImage"`
	KernelVersion      string            `json:"kernelVersion"`
	ContainerRuntime   string            `json:"containerRuntime"`
	KubeletVersion     string            `json:"kubeletVersion"`
	KubeProxyVersion   string            `json:"kubeProxyVersion"`
	SystemUUID         string            `json:"systemUUID"`
	CpuCapacity        string            `json:"cpuCapacity"`
	CpuAllocatable     string            `json:"cpuAllocatable"`
	MemoryCapacity     string            `json:"memoryCapacity"`
	MemoryAllocatable  string            `json:"memoryAllocatable"`
	StorageCapacity    string            `json:"storageCapacity"`
	StorageAllocatable string            `json:"storageAllocatable"`
	MaxPods            string            `json:"maxPods"`
	NetworkPolicy      string            `json:"networkPolicy"`
	Labels             map[string]string `json:"labels"`
	Annotations        map[string]string `json:"annotations"`
	Cordoned           bool              `json:"cordoned"`
}

// ComponentInfo 组件信息
type ComponentInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ResourceStatus 资源状态
type ResourceStatus struct {
	TotalNodes   int `json:"totalNodes"`
	HealthyNodes int `json:"healthyNodes"`
	WarningPods  int `json:"warningPods"`
	FailedPods   int `json:"failedPods"`
	PendingPods  int `json:"pendingPods"`
}
