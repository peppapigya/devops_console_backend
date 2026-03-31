package chaos

import (
	"context"
	"fmt"
	"time"

	"devops-console-backend/internal/dal/request/k8s"
	chaosStrategies "devops-console-backend/internal/services/k8s/chaos/strategies"
	"devops-console-backend/pkg/configs"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ChaosService struct{}

func NewChaosService() *ChaosService {
	return &ChaosService{}
}

// ChaosFactory 混沌实验工厂
type ChaosFactory struct {
	strategies map[string]chaosStrategies.FaultStrategy
}

// NewChaosFactory 创建混沌实验工厂
func NewChaosFactory() *ChaosFactory {
	return &ChaosFactory{
		strategies: map[string]chaosStrategies.FaultStrategy{
			"PodChaos":     &chaosStrategies.PodChaosStrategy{},
			"NetworkChaos": &chaosStrategies.NetworkChaosStrategy{},
			"IOChaos":      &chaosStrategies.IOChaosStrategy{},
			"StressChaos":  &chaosStrategies.StressChaosStrategy{},
		},
	}
}

// CreateExperiment 创建混沌实验
func (f *ChaosFactory) CreateExperiment(faultType string, request interface{}) (*unstructured.Unstructured, error) {
	strategy, exists := f.strategies[faultType]
	if !exists {
		return nil, fmt.Errorf("unsupported fault type: %s", faultType)
	}

	return strategy.CreateSpec(request)
}

// GetStrategy 获取策略
func (f *ChaosFactory) GetStrategy(faultType string) (chaosStrategies.FaultStrategy, error) {
	strategy, exists := f.strategies[faultType]
	if !exists {
		return nil, fmt.Errorf("unsupported fault type: %s", faultType)
	}
	return strategy, nil
}

// getResourceKind 获取资源类型
func getResourceKind(faultType string) string {
	switch faultType {
	case "PodChaos":
		return "podchaos"
	case "NetworkChaos":
		return "networkchaos"
	case "IOChaos":
		return "ioschaos"
	case "StressChaos":
		return "stresses"
	default:
		return faultType + "s"
	}
}

// convertTime 转换时间
func convertTime(t interface{}) *time.Time {
	if t == nil {
		return nil
	}
	if timeVal, ok := t.(time.Time); ok {
		return &timeVal
	}
	return nil
}

// CreateChaosExperiment 创建混沌实验（模板方法）
func (s *ChaosService) CreateChaosExperiment(ctx context.Context, instanceID uint, req *k8s.ChaosExperimentCreateRequest) error {
	dynamicClient, exists := configs.GetK8sDynamicClient(instanceID)
	if !exists {
		return fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}
	factory := NewChaosFactory()
	chaosObj, err := factory.CreateExperiment(req.FaultType, req)
	if err != nil {
		return fmt.Errorf("create chaos spec failed: %w", err)
	}

	gvk := schema.GroupVersionKind{
		Group:   "chaos-mesh.org",
		Version: "v1alpha1",
		Kind:    req.FaultType,
	}

	chaosObj.SetGroupVersionKind(gvk)

	gvr := schema.GroupVersionResource{
		Group:    "chaos-mesh.org",
		Version:  "v1alpha1",
		Resource: getResourceKind(req.FaultType),
	}

	_, err = dynamicClient.Resource(gvr).Namespace(req.Namespace).Create(ctx, chaosObj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create chaos experiment failed: %w", err)
	}

	return nil
}

// ListChaosExperiments 查询混沌实验列表（模板方法）
func (s *ChaosService) ListChaosExperiments(ctx context.Context, instanceID uint, req *k8s.ChaosExperimentListRequest) ([]k8s.ChaosExperimentListItem, error) {
	dynamicClient, exists := configs.GetK8sDynamicClient(instanceID)

	if !exists {
		return nil, fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	namespace := metav1.NamespaceAll
	var results []k8s.ChaosExperimentListItem
	if req.Namespace != "" && req.Namespace != "all" {
		namespace = req.Namespace
	}
	types := []string{"PodChaos", "NetworkChaos", "IOChaos", "StressChaos"}
	for _, t := range types {
		if req.Type != "" && req.Type != "all" && req.Type != t {
			continue
		}

		gvr := schema.GroupVersionResource{
			Group:    "chaos-mesh.org",
			Version:  "v1alpha1",
			Resource: getResourceKind(t),
		}
		list, err := dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}

		for _, item := range list.Items {
			experiment := s.convertToExperimentItem(&item, t)
			if req.Status != "" && experiment.Status != req.Status {
				continue
			}
			results = append(results, experiment)
		}
	}

	return results, nil
}

// GetChaosExperiment 获取混沌实验详情
func (s *ChaosService) GetChaosExperiment(ctx context.Context, instanceID uint, namespace, name string) (*k8s.ChaosExperimentDetail, error) {
	dynamicClient, exists := configs.GetK8sDynamicClient(instanceID)

	if !exists {
		return nil, fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	types := []string{"PodChaos", "NetworkChaos", "IOChaos", "StressChaos"}
	for _, t := range types {
		gvr := schema.GroupVersionResource{
			Group:    "chaos-mesh.org",
			Version:  "v1alpha1",
			Resource: getResourceKind(t),
		}

		obj, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			continue
		}

		detail := s.convertToExperimentDetail(obj, t)
		return &detail, nil
	}

	return nil, fmt.Errorf("chaos experiment not found")
}

// DeleteChaosExperiment 删除混沌实验（模板方法）
func (s *ChaosService) DeleteChaosExperiment(ctx context.Context, instanceID uint, namespace, name string) error {
	dynamicClient, exists := configs.GetK8sDynamicClient(instanceID)

	if !exists {
		return fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	types := []string{"PodChaos", "NetworkChaos", "IOChaos", "StressChaos"}
	for _, t := range types {
		gvr := schema.GroupVersionResource{
			Group:    "chaos-mesh.org",
			Version:  "v1alpha1",
			Resource: getResourceKind(t),
		}

		err := dynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			continue
		}

		return nil
	}

	return fmt.Errorf("chaos experiment not found")
}

// PauseChaosExperiment 暂停混沌实验
func (s *ChaosService) PauseChaosExperiment(ctx context.Context, instanceID uint, namespace, name string) error {
	return s.setExperimentStatus(ctx, instanceID, namespace, name, true)
}

// ResumeChaosExperiment 恢复混沌实验
func (s *ChaosService) ResumeChaosExperiment(ctx context.Context, instanceID uint, namespace, name string) error {
	return s.setExperimentStatus(ctx, instanceID, namespace, name, false)
}

// setExperimentStatus 设置实验状态（模板方法）
func (s *ChaosService) setExperimentStatus(ctx context.Context, instanceID uint, namespace, name string, paused bool) error {
	dynamicClient, exists := configs.GetK8sDynamicClient(instanceID)

	if !exists {
		return fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	types := []string{"PodChaos", "NetworkChaos", "IOChaos", "StressChaos"}
	for _, t := range types {
		gvr := schema.GroupVersionResource{
			Group:    "chaos-mesh.org",
			Version:  "v1alpha1",
			Resource: getResourceKind(t),
		}

		obj, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			continue
		}

		annotations := obj.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}

		if paused {
			annotations["chaos-mesh.org/pause"] = "true"
		} else {
			delete(annotations, "chaos-mesh.org/pause")
		}

		obj.SetAnnotations(annotations)

		_, err = dynamicClient.Resource(gvr).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("update chaos experiment failed: %w", err)
		}

		return nil
	}

	return fmt.Errorf("chaos experiment not found")
}

// convertToExperimentItem 转换为实验列表项
func (s *ChaosService) convertToExperimentItem(obj *unstructured.Unstructured, faultType string) k8s.ChaosExperimentListItem {
	annotations := obj.GetAnnotations()
	labels := obj.GetLabels()
	fmt.Printf("object%v\n", obj.Object)
	status := getChaosPhase(obj)

	fmt.Printf("获取的状态%v,获取到的：%v", status, s.getStatus(annotations, status))
	return k8s.ChaosExperimentListItem{
		Name:        obj.GetName(),
		Namespace:   obj.GetNamespace(),
		FaultType:   faultType,
		Status:      s.getStatus(annotations, status),
		Phase:       status,
		CreatedAt:   obj.GetCreationTimestamp().Time,
		Labels:      labels,
		Annotations: annotations,
	}
}

func getChaosPhase(obj *unstructured.Unstructured) string {
	condMap := map[string]string{}
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err == nil && found {
		for _, c := range conditions {
			m, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			t, _, _ := unstructured.NestedString(m, "type")
			s, _, _ := unstructured.NestedString(m, "status")
			if t != "" {
				condMap[t] = s
			}
		}
	}
	action, _, _ := unstructured.NestedString(obj.Object, "spec", "action")
	switch {
	case condMap["Paused"] == "True":
		return "Paused"
	case condMap["AllRecovered"] == "True":
		return "Finished"
	case action == "pod-kill" && condMap["AllInjected"] == "True":
		return "Executed"
	case condMap["AllInjected"] == "True":
		return "Running"
	case condMap["Selected"] == "True":
		return "Selected"
	default:
		return "Unknown"
	}
}

// convertToExperimentDetail 转换为实验详情
func (s *ChaosService) convertToExperimentDetail(obj *unstructured.Unstructured, faultType string) k8s.ChaosExperimentDetail {
	annotations := obj.GetAnnotations()
	labels := obj.GetLabels()
	status, _, _ := unstructured.NestedString(obj.Object, "status", "phase")

	startedAt, _, _ := unstructured.NestedFieldNoCopy(obj.Object, "status", "experiment", "startTime")
	finishedAt, _, _ := unstructured.NestedFieldNoCopy(obj.Object, "status", "experiment", "endTime")

	duration, _, _ := unstructured.NestedString(obj.Object, "spec", "duration")
	spec, found, err := unstructured.NestedMap(obj.Object, "spec")

	if err != nil || !found {
		spec = make(map[string]interface{})
	}

	detail := k8s.ChaosExperimentDetail{
		Name:        obj.GetName(),
		Namespace:   obj.GetNamespace(),
		FaultType:   faultType,
		Status:      s.getStatus(annotations, status),
		Phase:       status,
		CreatedAt:   obj.GetCreationTimestamp().Time,
		StartedAt:   convertTime(startedAt),
		FinishedAt:  convertTime(finishedAt),
		Duration:    duration,
		Labels:      labels,
		Annotations: annotations,
		Spec:        spec,
		Events:      s.extractEvents(obj),
		Targets:     s.extractTargets(obj),
	}

	return detail
}

// getStatus 获取状态
func (s *ChaosService) getStatus(annotations map[string]string, phase string) string {
	if paused, ok := annotations["chaos-mesh.org/pause"]; ok && paused == "true" {
		return "paused"
	}
	return phase
}

// extractEvents 提取事件
func (s *ChaosService) extractEvents(obj *unstructured.Unstructured) []k8s.EventInfo {
	return []k8s.EventInfo{}
}

// extractTargets 提取目标
func (s *ChaosService) extractTargets(obj *unstructured.Unstructured) []k8s.TargetInfo {
	return []k8s.TargetInfo{}
}
