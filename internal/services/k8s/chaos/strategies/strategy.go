package strategies

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// FaultStrategy 故障注入策略接口
type FaultStrategy interface {
	CreateSpec(request interface{}) (*unstructured.Unstructured, error)
	GetGVK() schema.GroupVersionKind
}
