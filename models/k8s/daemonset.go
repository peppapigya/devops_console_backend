package k8s

import "time"

// DaemonSetCreateRequest 创建DaemonSet请求
type DaemonSetCreateRequest struct {
	Name   string            `json:"name" binding:"required"`
	Labels map[string]string `json:"labels"`
	Image  string            `json:"image" binding:"required"`
	Port   int32             `json:"port"`
}

// DaemonSetUpdateRequest 更新DaemonSet请求
type DaemonSetUpdateRequest struct {
	Image string `json:"image" binding:"required"`
}

// DaemonSetListItem DaemonSet列表项
type DaemonSetListItem struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Current   int32     `json:"current"`
	Desired   int32     `json:"desired"`
	Ready     int32     `json:"ready"`
	Available int32     `json:"available"`
	Created   time.Time `json:"created"`
}

// DaemonSetDetail DaemonSet详情
type DaemonSetDetail struct {
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	Current    int32                  `json:"current"`
	Desired    int32                  `json:"desired"`
	Ready      int32                  `json:"ready"`
	Available  int32                  `json:"available"`
	Conditions []interface{}          `json:"conditions"`
	Labels     map[string]string      `json:"labels"`
	Age        int64                  `json:"age"`
}