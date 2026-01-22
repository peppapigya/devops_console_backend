package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ServicePortRequest 服务端口请求
type ServicePortRequest struct {
	Name       string             `json:"name"`
	Port       int32              `json:"port"`
	TargetPort intstr.IntOrString `json:"targetPort"`
	NodePort   int32              `json:"nodePort"`
	Protocol   string             `json:"protocol"`
}

// ServiceCreateRequest 创建Service请求
type ServiceCreateRequest struct {
	Labels   map[string]string    `json:"labels"`
	Type     corev1.ServiceType   `json:"type"`
	Selector map[string]string    `json:"selector"`
	Ports    []ServicePortRequest `json:"ports"`
}

// ServiceUpdateRequest 更新Service请求
type ServiceUpdateRequest struct {
	Labels   map[string]string    `json:"labels"`
	Selector map[string]string    `json:"selector"`
	Ports    []ServicePortRequest `json:"ports"`
}

// ServiceListItem Service列表项
type ServiceListItem struct {
	Namespace  string               `json:"namespace"`
	Name       string               `json:"name"`
	Type       corev1.ServiceType   `json:"type"`
	ClusterIP  string               `json:"clusterIP"`
	ExternalIP []string             `json:"externalIP"`
	Selector   map[string]string    `json:"selector"`
	Ports      []corev1.ServicePort `json:"ports"`
	CreatedAt  string               `json:"createdAt"`
	Labels     map[string]string    `json:"labels"`
	Age        int64                `json:"age"`
}

// ServiceDetail Service详情
type ServiceDetail struct {
	Name                  string                                  `json:"name"`
	Namespace             string                                  `json:"namespace"`
	Status                corev1.ServiceStatus                    `json:"status"`
	Labels                map[string]string                       `json:"labels"`
	Annotations           map[string]string                       `json:"annotations"`
	Selector              map[string]string                       `json:"selector"`
	Type                  corev1.ServiceType                      `json:"type"`
	IP                    string                                  `json:"ip"`
	IPs                   []string                                `json:"ips"`
	Ports                 []corev1.ServicePort                    `json:"ports"`
	SessionAffinity       corev1.ServiceAffinity                  `json:"sessionAffinity"`
	ExternalTrafficPolicy corev1.ServiceExternalTrafficPolicyType `json:"externalTrafficPolicy"`
	Events                []corev1.LoadBalancerIngress            `json:"events"`
	Endpoints             interface{}                             `json:"endpoints"`
	Age                   int64                                   `json:"age"`
}

// MultipleDeleteRequest 批量删除请求
type MultipleDeleteRequest struct {
	Services []string `json:"services" binding:"required"`
}