package k8s

import "time"

// NamespaceListItem Namespace列表项
type NamespaceListItem struct {
	Name              string            `json:"name"`
	Status            string            `json:"status"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Age               int64             `json:"age"`
}

// NamespaceDetail Namespace详情
type NamespaceDetail struct {
	Name              string            `json:"name"`
	Status            string            `json:"status"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Age               int64             `json:"age"`
}
