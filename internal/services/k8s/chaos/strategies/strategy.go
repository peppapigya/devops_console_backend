package strategies

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// FaultStrategy 故障注入策略接口
type FaultStrategy interface {
	// CreateSpec 创建故障注入策略
	CreateSpec(request interface{}) (*unstructured.Unstructured, error)
	// GetGVK 获取故障注入策略的GVK
	GetGVK() schema.GroupVersionKind
}
