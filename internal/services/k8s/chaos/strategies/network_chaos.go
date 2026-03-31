package strategies

import (
	"fmt"

	"devops-console-backend/internal/dal/request/k8s"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NetworkChaosStrategy 网络混沌策略
type NetworkChaosStrategy struct{}

func (s *NetworkChaosStrategy) GetGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "chaos-mesh.org",
		Version: "v1alpha1",
		Kind:    "NetworkChaos",
	}
}

func (s *NetworkChaosStrategy) CreateSpec(request interface{}) (*unstructured.Unstructured, error) {
	req, ok := request.(*k8s.ChaosExperimentCreateRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for NetworkChaos")
	}

	spec, ok := req.Spec.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec type for NetworkChaos")
	}

	selector := buildSelector(spec["selector"])
	networkSpec := map[string]interface{}{
		"action":    spec["action"],
		"selector":  selector,
		"duration":  req.Duration,
		"direction": spec["direction"],
	}

	if delay, ok := spec["delay"].(map[string]interface{}); ok {
		networkSpec["delay"] = delay
	}
	if loss, ok := spec["loss"].(map[string]interface{}); ok {
		networkSpec["loss"] = loss
	}
	if duplicate, ok := spec["duplicate"].(map[string]interface{}); ok {
		networkSpec["duplicate"] = duplicate
	}
	if corrupt, ok := spec["corrupt"].(map[string]interface{}); ok {
		networkSpec["corrupt"] = corrupt
	}
	if externalTargets, ok := spec["externalTargets"].([]interface{}); ok && len(externalTargets) > 0 {
		networkSpec["externalTargets"] = externalTargets
	}

	chaos := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "chaos-mesh.org/v1alpha1",
			"kind":       "NetworkChaos",
			"metadata": map[string]interface{}{
				"name":        req.Name,
				"namespace":   req.Namespace,
				"labels":      req.Labels,
				"annotations": req.Annotations,
			},
			"spec": networkSpec,
		},
	}

	return chaos, nil
}
