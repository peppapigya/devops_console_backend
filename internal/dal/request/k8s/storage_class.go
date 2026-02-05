package k8s

import v1 "k8s.io/api/core/v1"

// StorageClassListItem SC列表项
type StorageClassListItem struct {
	Name              string                            `json:"name"`
	Provisioner       string                            `json:"provisioner"`
	ReclaimPolicy     *v1.PersistentVolumeReclaimPolicy `json:"reclaimPolicy"`
	VolumeBindingMode string                            `json:"volumeBindingMode"`
	Age               int64                             `json:"age"`
}

type StorageClassResponse struct {
	Name              string     `json:"name"`
	Age               int64      `json:"age"`
	Namespace         string     `json:"namespace"`
	CreationTimestamp string     `json:"creationTimestamp"`
	Labels            []KeyValue `json:"labels"`
	Annotations       []KeyValue `json:"annotations"`
	Provisioner       string     `json:"provisioner"`
	ContinueToken     string     `json:"continueToken"`
}

// StorageClassCreateRequest 创建SC请求
type StorageClassCreateRequest struct {
	Name          string     `json:"name" binding:"required"`
	Provisioner   string     `json:"provisioner" binding:"required"`
	ReclaimPolicy string     `json:"reclaimPolicy"`
	Parameters    []KeyValue `json:"parameters"`
	YAML          string     `json:"yaml"`
}
