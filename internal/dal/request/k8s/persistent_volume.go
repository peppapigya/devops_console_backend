package k8s

import corev1 "k8s.io/api/core/v1"

// PersistentVolumeListItem PV列表项
type PersistentVolumeListItem struct {
	Name          string                              `json:"name"`
	Status        string                              `json:"status"`
	Claim         string                              `json:"claim"`
	AccessModes   []corev1.PersistentVolumeAccessMode `json:"accessModes"`
	Capacity      string                              `json:"capacity"`
	StorageClass  string                              `json:"storageClass"`
	ReclaimPolicy string                              `json:"reclaimPolicy"`
	Reason        string                              `json:"reason"`
	Age           int64                               `json:"age"`
}

// PersistentVolumeCreateRequest 创建PV请求
type PersistentVolumeCreateRequest struct {
	Name             string           `json:"name" binding:"required"`
	StorageClassName string           `json:"storageClassName"`
	Capacity         string           `json:"capacity" binding:"required"`
	AccessModes      []string         `json:"accessModes" binding:"required"`
	HostPath         string           `json:"hostPath"`
	NFS              *NFSVolumeSource `json:"nfs"`
	Labels           []KeyValue       `json:"labels"`
	YAML             string           `json:"yaml"`
}

// NFSVolumeSource NFS配置
type NFSVolumeSource struct {
	Server string `json:"server"`
	Path   string `json:"path"`
}
