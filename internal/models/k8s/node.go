package k8s

// NodeListItem 节点列表项
type NodeListItem struct {
	Name              string            `json:"name"`
	Status            string            `json:"status"`
	Roles             string            `json:"roles"`
	InternalIP        string            `json:"internalIP"`
	ExternalIP        string            `json:"externalIP"`
	Version           string            `json:"version"`
	Age               int64             `json:"age"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	Cordoned          bool              `json:"cordoned"`
	CPUCapacity       string            `json:"cpuCapacity"`
	CPUAllocatable    string            `json:"cpuAllocatable"`
	MemoryCapacity    string            `json:"memoryCapacity"`
	MemoryAllocatable string            `json:"memoryAllocatable"`
	StorageCapacity   string            `json:"storageCapacity"`
	PodCapacity       string            `json:"podCapacity"`
	PodCount          int               `json:"podCount"`
	OSImage           string            `json:"osImage"`
	KernelVersion     string            `json:"kernelVersion"`
	ContainerRuntime  string            `json:"containerRuntime"`
	KubeletVersion    string            `json:"kubeletVersion"`
	KubeProxyVersion  string            `json:"kubeProxyVersion"`
	SystemUUID        string            `json:"systemUUID"`
	CreateTime        int64             `json:"createTime"`
}

// NodeDetail 节点详情
type NodeDetail struct {
	Name              string            `json:"name"`
	UID               string            `json:"uid"`
	CreationTimestamp int64             `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	Status            string            `json:"status"`
	Conditions        []NodeCondition   `json:"conditions"`
	Addresses         []NodeAddress     `json:"addresses"`
	NodeInfo          NodeSystemInfo    `json:"nodeInfo"`
	Capacity          map[string]string `json:"capacity"`
	Allocatable       map[string]string `json:"allocatable"`
	Pods              []PodOnNode       `json:"pods"`
}

// NodeCondition 节点条件
type NodeCondition struct {
	Type           string `json:"type"`
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	Message        string `json:"message"`
	LastHeartbeat  int64  `json:"lastHeartbeat"`
	LastTransition int64  `json:"lastTransition"`
}

// NodeAddress 节点地址
type NodeAddress struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

// NodeSystemInfo 节点系统信息
type NodeSystemInfo struct {
	MachineID               string `json:"machineID"`
	SystemUUID              string `json:"systemUUID"`
	BootID                  string `json:"bootID"`
	KernelVersion           string `json:"kernelVersion"`
	OSImage                 string `json:"osImage"`
	ContainerRuntimeVersion string `json:"containerRuntimeVersion"`
	KubeletVersion          string `json:"kubeletVersion"`
	KubeProxyVersion        string `json:"kubeProxyVersion"`
	OperatingSystem         string `json:"operatingSystem"`
	Architecture            string `json:"architecture"`
}

// PodOnNode 节点上的Pod
type PodOnNode struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Created   int64  `json:"created"`
}
