package k8s

import (
	"time"
)

// DeploymentCreateRequest 创建Deployment请求
type DeploymentCreateRequest struct {
	Name     string            `json:"name" binding:"required"`
	Replicas int32             `json:"replicas" binding:"required"`
	Labels   map[string]string `json:"labels"`
	Image    string            `json:"image" binding:"required"`
	Port     int32             `json:"port"`
}

// DeploymentUpdateRequest 更新Deployment请求
type DeploymentUpdateRequest struct {
	Image string `json:"image" binding:"required"`
}

// ContainerInfo 容器信息
type ContainerInfo struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// ResourceInfo 资源信息
type ResourceInfo struct {
	CPURequest    string `json:"cpuRequest"`
	CPULimit      string `json:"cpuLimit"`
	MemoryRequest string `json:"memoryRequest"`
	MemoryLimit   string `json:"memoryLimit"`
}

// DeploymentListItem Deployment列表项
type DeploymentListItem struct {
	Name      string       `json:"name"`
	Namespace string       `json:"namespace"`
	Replicas  int32        `json:"replicas"`
	Ready     int32        `json:"ready"`
	Available int32        `json:"available"`
	Created   time.Time    `json:"created"`
	Image     string       `json:"image"`
	Resources ResourceInfo `json:"resources"`
}

// DeploymentDetail Deployment详情
type DeploymentDetail struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Replicas   int32             `json:"replicas"`
	Ready      int32             `json:"ready"`
	Available  int32             `json:"available"`
	Conditions []interface{}     `json:"conditions"`
	Labels     map[string]string `json:"labels"`
	Selector   map[string]string `json:"selector"`
	Age        int64             `json:"age"`
	Containers []ContainerInfo   `json:"containers"`
}
