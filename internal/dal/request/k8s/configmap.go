package k8s

// ConfigMapListItem ConfigMap列表项
type ConfigMapListItem struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	DataCount int    `json:"dataCount"`
	Age       int64  `json:"age"`
}

// ConfigMapCreateRequest 创建ConfigMap请求
type ConfigMapCreateRequest struct {
	Name      string            `json:"name" binding:"required"`
	Namespace string            `json:"namespace" binding:"required"`
	Data      map[string]string `json:"data"`
	YAML      string            `json:"yaml"`
}

// ConfigMapUpdateRequest 更新ConfigMap请求
type ConfigMapUpdateRequest struct {
	Data map[string]string `json:"data"`
	YAML string            `json:"yaml"`
}
