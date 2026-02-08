package k8s

// SecretListItem Secret列表项
type SecretListItem struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	DataCount int    `json:"dataCount"`
	Age       int64  `json:"age"`
}

// SecretCreateRequest 创建Secret请求
type SecretCreateRequest struct {
	Name      string            `json:"name" binding:"required"`
	Namespace string            `json:"namespace" binding:"required"`
	Type      string            `json:"type"` // Opaque, kubernetes.io/tls, etc.
	Data      map[string]string `json:"data"` // base64 encoded values
	YAML      string            `json:"yaml"`
}

// SecretUpdateRequest 更新Secret请求
type SecretUpdateRequest struct {
	Data map[string]string `json:"data"` // base64 encoded values
	YAML string            `json:"yaml"`
}
