package strategies

import (
	"fmt"

	"devops-console-backend/internal/dal/request/k8s"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PodChaosStrategy Pod混沌策略
type PodChaosStrategy struct{}

func (s *PodChaosStrategy) GetGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "chaos-mesh.org",
		Version: "v1alpha1",
		Kind:    "PodChaos",
	}
}

func (s *PodChaosStrategy) CreateSpec(request interface{}) (*unstructured.Unstructured, error) {
	req, ok := request.(*k8s.ChaosExperimentCreateRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for PodChaos")
	}

	spec, ok := req.Spec.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec type for PodChaos")
	}

	selector := buildSelector(spec["selector"])

	chaos := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "chaos-mesh.org/v1alpha1",
			"kind":       "PodChaos",
			"metadata": map[string]interface{}{
				"name":        req.Name,
				"namespace":   req.Namespace,
				"labels":      req.Labels,
				"annotations": req.Annotations,
			},
			"spec": map[string]interface{}{
				"action":      spec["action"],
				"mode":        spec["mode"],
				"value":       spec["value"],
				"selector":    selector,
				"duration":    req.Duration,
				"gracePeriod": spec["gracePeriod"],
			},
		},
	}

	if containerName, ok := spec["containerName"].(string); ok && containerName != "" {
		chaos.Object["spec"].(map[string]interface{})["containerName"] = containerName
	}

	return chaos, nil
}
