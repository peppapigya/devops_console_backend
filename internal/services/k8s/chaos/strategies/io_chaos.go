package strategies

import (
	"fmt"

	"devops-console-backend/internal/dal/request/k8s"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IOChaosStrategy IO混沌策略
type IOChaosStrategy struct{}

func (s *IOChaosStrategy) GetGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "chaos-mesh.org",
		Version: "v1alpha1",
		Kind:    "IOChaos",
	}
}

func (s *IOChaosStrategy) CreateSpec(request interface{}) (*unstructured.Unstructured, error) {
	req, ok := request.(*k8s.ChaosExperimentCreateRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for IOChaos")
	}

	spec, ok := req.Spec.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec type for IOChaos")
	}

	selector := buildSelector(spec["selector"])

	ioSpec := map[string]interface{}{
		"action":   spec["action"],
		"selector": selector,
		"duration": req.Duration,
		"percent":  spec["percent"],
		"path":     spec["path"],
		"methods":  spec["methods"],
	}

	if faults, ok := spec["faults"].([]interface{}); ok && len(faults) > 0 {
		ioSpec["faults"] = faults
	}

	chaos := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "chaos-mesh.org/v1alpha1",
			"kind":       "IOChaos",
			"metadata": map[string]interface{}{
				"name":        req.Name,
				"namespace":   req.Namespace,
				"labels":      req.Labels,
				"annotations": req.Annotations,
			},
			"spec": ioSpec,
		},
	}

	return chaos, nil
}
