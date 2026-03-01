package k8s

import (
	"time"
)

// StatefulSetCreateRequest 创建StatefulSet请求
type StatefulSetCreateRequest struct {
	Name     string            `json:"name" binding:"required"`
	Replicas int32             `json:"replicas" binding:"required"`
	Labels   map[string]string `json:"labels"`
	Image    string            `json:"image" binding:"required"`
	Port     int32             `json:"port"`
}

// StatefulSetUpdateRequest 更新StatefulSet请求
type StatefulSetUpdateRequest struct {
	Image string `json:"image" binding:"required"`
}

// StatefulSetListItem StatefulSet列表项
type StatefulSetListItem struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Replicas  int32             `json:"replicas"`
	Ready     int32             `json:"ready"`
	Created   time.Time         `json:"created"`
	Image     string            `json:"image"`
	Resources ResourceInfo      `json:"resources"`
	Labels    map[string]string `json:"labels"`
}

// StatefulSetDetail StatefulSet详情
type StatefulSetDetail struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Replicas   int32             `json:"replicas"`
	Ready      int32             `json:"ready"`
	Conditions []interface{}     `json:"conditions"`
	Labels     map[string]string `json:"labels"`
	Selector   map[string]string `json:"selector"`
	Age        int64             `json:"age"`
	Containers []ContainerInfo   `json:"containers"`
}
