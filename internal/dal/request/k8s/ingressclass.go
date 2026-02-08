package k8s

// IngressClassListItem IngressClass列表项
type IngressClassListItem struct {
	Name       string  `json:"name"`
	Controller string  `json:"controller"`
	IsDefault  bool    `json:"isDefault"`
	Parameters *string `json:"parameters"`
	Age        int64   `json:"age"`
}
