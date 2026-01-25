package k8s

// PodListItem Pod列表项
type PodListItem struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     string `json:"ready"`
	Status    string `json:"status"`
	Restarts  int32  `json:"restarts"`
	Age       int64  `json:"age"`
	IP        string `json:"ip"`
	Node      string `json:"node"`
}

// ContainerPort 容器端口
type ContainerPort struct {
	Name          string `json:"name"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol"`
}

// ContainerEnv 容器环境变量
type ContainerEnv struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ResourceRequirements 资源需求
type ResourceRequirements struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// ContainerResources 容器资源
type ContainerResources struct {
	Requests ResourceRequirements `json:"requests"`
	Limits   ResourceRequirements `json:"limits"`
}

// Container 容器配置
type Container struct {
	Name            string             `json:"name"`
	Image           string             `json:"image"`
	ImagePullPolicy string             `json:"imagePullPolicy"`
	Ports           []ContainerPort    `json:"ports"`
	Env             []ContainerEnv     `json:"env"`
	Resources       ContainerResources `json:"resources"`
}

// KeyValue 键值对
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PodCreateRequest 创建Pod请求
type PodCreateRequest struct {
	Podname       string      `json:"podname" binding:"required"`
	Namespace     string      `json:"namespace" binding:"required"`
	RestartPolicy string      `json:"restartPolicy"`
	Labels        []KeyValue  `json:"labels"`
	Annotations   []KeyValue  `json:"annotations"`
	NodeSelector  []KeyValue  `json:"nodeSelector"`
	Containers    []Container `json:"containers"`
	YAML          string      `json:"yaml"`
}

// PodUpdateRequest 更新Pod请求
type PodUpdateRequest struct {
	Podname   string `json:"podname" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Imagename string `json:"imagename" binding:"required"`
	Image     string `json:"image" binding:"required"`
}
