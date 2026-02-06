package k8s

// IngressListItem Ingress列表项
type IngressListItem struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	ClassName *string  `json:"className"`
	Hosts     []string `json:"hosts"`
	Address   string   `json:"address"`
	Age       int64    `json:"age"`
}

// IngressRule Ingress规则
type IngressRule struct {
	Host  string            `json:"host"`
	Paths []IngressHttpPath `json:"paths"`
}

// IngressHttpPath HTTP路径
type IngressHttpPath struct {
	Path        string `json:"path"`
	PathType    string `json:"pathType"`
	ServiceName string `json:"serviceName"`
	ServicePort int32  `json:"servicePort"`
}

// IngressTLS TLS配置
type IngressTLS struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secretName"`
}

// IngressCreateRequest 创建Ingress请求
type IngressCreateRequest struct {
	Name        string            `json:"name" binding:"required"`
	Namespace   string            `json:"namespace" binding:"required"`
	ClassName   *string           `json:"className"`
	Rules       []IngressRule     `json:"rules"`
	TLS         []IngressTLS      `json:"tls"`
	Annotations map[string]string `json:"annotations"`
	YAML        string            `json:"yaml"`
}

// IngressUpdateRequest 更新Ingress请求
type IngressUpdateRequest struct {
	ClassName   *string           `json:"className"`
	Rules       []IngressRule     `json:"rules"`
	TLS         []IngressTLS      `json:"tls"`
	Annotations map[string]string `json:"annotations"`
	YAML        string            `json:"yaml"`
}
