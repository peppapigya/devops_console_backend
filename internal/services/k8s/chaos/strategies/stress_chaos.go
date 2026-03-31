package strategies

import (
	"fmt"

	"devops-console-backend/internal/dal/request/k8s"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// StressChaosStrategy 压力混沌策略
type StressChaosStrategy struct{}

func (s *StressChaosStrategy) GetGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "chaos-mesh.org",
		Version: "v1alpha1",
		Kind:    "StressChaos",
	}
}

func (s *StressChaosStrategy) CreateSpec(request interface{}) (*unstructured.Unstructured, error) {
	req, ok := request.(*k8s.ChaosExperimentCreateRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for StressChaos")
	}

	spec, ok := req.Spec.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec type for StressChaos")
	}

	selector := buildSelector(spec["selector"])

	stressSpec := map[string]interface{}{
		"selector": selector,
		"duration": req.Duration,
		"workers":  spec["workers"],
	}

	if cpu, ok := spec["cpu"].(map[string]interface{}); ok {
		stressSpec["cpu"] = cpu
	}
	if memory, ok := spec["memory"].(map[string]interface{}); ok {
		stressSpec["memory"] = memory
	}

	chaos := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "chaos-mesh.org/v1alpha1",
			"kind":       "StressChaos",
			"metadata": map[string]interface{}{
				"name":        req.Name,
				"namespace":   req.Namespace,
				"labels":      req.Labels,
				"annotations": req.Annotations,
			},
			"spec": stressSpec,
		},
	}

	return chaos, nil
}
