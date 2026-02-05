package k8s

import corev1 "k8s.io/api/core/v1"

// PersistentVolumeClaimListItem PVC列表项
type PersistentVolumeClaimListItem struct {
	Name         string                              `json:"name"`
	Namespace    string                              `json:"namespace"`
	Status       string                              `json:"status"`
	Volume       string                              `json:"volume"`
	Capacity     string                              `json:"capacity"`
	AccessModes  []corev1.PersistentVolumeAccessMode `json:"accessModes"`
	StorageClass *string                             `json:"storageClass"`
	Age          int64                               `json:"age"`
}

// PersistentVolumeClaimCreateRequest 创建PVC请求
type PersistentVolumeClaimCreateRequest struct {
	Name             string     `json:"name" binding:"required"`
	Namespace        string     `json:"namespace" binding:"required"`
	StorageClassName string     `json:"storageClassName"`
	AccessModes      []string   `json:"accessModes" binding:"required"`
	Capacity         string     `json:"capacity" binding:"required"`
	Labels           []KeyValue `json:"labels"`
	YAML             string     `json:"yaml"`
}
