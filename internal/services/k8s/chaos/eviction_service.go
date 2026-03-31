package chaos

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	reqK8s "devops-console-backend/internal/dal/request/k8s"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils/logs"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EvictionService 演练节点驱逐服务
type EvictionService struct{}

func NewEvictionService() *EvictionService {
	return &EvictionService{}
}

// GetChaosTestingNodes 获取所有带 role=chaos-testing 标签的演练节点
func (s *EvictionService) GetChaosTestingNodes(ctx context.Context, instanceID uint) ([]reqK8s.ChaosNode, error) {
	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		return nil, fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: "role=chaos-testing",
	})
	if err != nil {
		return nil, fmt.Errorf("获取演练节点失败: %w", err)
	}

	var result []reqK8s.ChaosNode
	for _, node := range nodeList.Items {
		status := "NotReady"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}
		result = append(result, reqK8s.ChaosNode{
			Name:   node.Name,
			Labels: node.Labels,
			Status: status,
		})
	}
	return result, nil
}

// TaintChaosNode 给演练节点打上 chaos-testing=true:NoSchedule 污点
func (s *EvictionService) TaintChaosNode(ctx context.Context, instanceID uint, nodeName string) error {
	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		return fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取节点失败: %w", err)
	}

	// 检查是否已有该 Taint
	for _, t := range node.Spec.Taints {
		if t.Key == "chaos-testing" {
			logs.Info(map[string]interface{}{"node": nodeName}, "Taint already exists, skip")
			return nil
		}
	}

	node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
		Key:    "chaos-testing",
		Value:  "true",
		Effect: corev1.TaintEffectNoSchedule,
	})

	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("打 Taint 失败: %w", err)
	}

	logs.Info(map[string]interface{}{"node": nodeName}, "成功给演练节点打 Taint")
	return nil
}

// UnTaintChaosNode 去掉演练节点的 chaos-testing Taint
func (s *EvictionService) UnTaintChaosNode(ctx context.Context, instanceID uint, nodeName string) error {
	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		return fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取节点失败: %w", err)
	}

	filtered := node.Spec.Taints[:0]
	for _, t := range node.Spec.Taints {
		if t.Key != "chaos-testing" {
			filtered = append(filtered, t)
		}
	}
	node.Spec.Taints = filtered

	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("去 Taint 失败: %w", err)
	}

	logs.Info(map[string]interface{}{"node": nodeName}, "成功去掉演练节点 Taint")
	return nil
}

// patchSpec is the structure used to patch Deployment spec
type deploymentPatch struct {
	Spec deploymentSpecPatch `json:"spec"`
}

type deploymentSpecPatch struct {
	Template podTemplatePatch `json:"template"`
}

type podTemplatePatch struct {
	Spec podSpecPatch `json:"spec"`
}

type podSpecPatch struct {
	Tolerations []corev1.Toleration `json:"tolerations"`
	Affinity    *corev1.Affinity    `json:"affinity"`
}

// PatchDeploymentForChaos 给 Deployment 打上 Toleration + NodeAffinity，使 Pod 调度到演练节点
// 返回原始 spec 的 JSON（用于后续回滚）
func (s *EvictionService) PatchDeploymentForChaos(ctx context.Context, instanceID uint, namespace, deploymentName, nodeName string) (string, error) {
	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		return "", fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	// 保存原始 tolerations 和 affinity
	origTolerations := deploy.Spec.Template.Spec.Tolerations
	origAffinity := deploy.Spec.Template.Spec.Affinity
	origSpec := podSpecPatch{
		Tolerations: origTolerations,
		Affinity:    origAffinity,
	}
	origJSON, err := json.Marshal(origSpec)
	if err != nil {
		return "", fmt.Errorf("序列化原始 spec 失败: %w", err)
	}

	// 构建新的 Toleration
	chaosToleration := corev1.Toleration{
		Key:      "chaos-testing",
		Operator: corev1.TolerationOpEqual,
		Value:    "true",
		Effect:   corev1.TaintEffectNoSchedule,
	}

	// 合并 tolerations（不重复添加）
	newTolerations := origTolerations
	hasToleration := false
	for _, t := range newTolerations {
		if t.Key == "chaos-testing" {
			hasToleration = true
			break
		}
	}
	if !hasToleration {
		newTolerations = append(newTolerations, chaosToleration)
	}

	// 构建 NodeAffinity（Required：只调度到指定演练节点）
	newAffinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/hostname",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{nodeName},
							},
						},
					},
				},
			},
		},
	}

	patch := deploymentPatch{
		Spec: deploymentSpecPatch{
			Template: podTemplatePatch{
				Spec: podSpecPatch{
					Tolerations: newTolerations,
					Affinity:    newAffinity,
				},
			},
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return "", fmt.Errorf("序列化 patch 失败: %w", err)
	}

	_, err = client.AppsV1().Deployments(namespace).Patch(ctx, deploymentName, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return "", fmt.Errorf("Patch Deployment 失败: %w", err)
	}

	logs.Info(map[string]interface{}{
		"deployment": deploymentName,
		"node":       nodeName,
	}, "成功 Patch Deployment 用于演练隔离")

	return string(origJSON), nil
}

// RestoreDeployment 将 Deployment 恢复到实验前的 spec
func (s *EvictionService) RestoreDeployment(ctx context.Context, instanceID uint, namespace, deploymentName, origSpecJSON string) error {
	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		return fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	var origSpec podSpecPatch
	if err := json.Unmarshal([]byte(origSpecJSON), &origSpec); err != nil {
		return fmt.Errorf("反序列化原始 spec 失败: %w", err)
	}

	patch := deploymentPatch{
		Spec: deploymentSpecPatch{
			Template: podTemplatePatch{
				Spec: origSpec,
			},
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("序列化恢复 patch 失败: %w", err)
	}

	_, err = client.AppsV1().Deployments(namespace).Patch(ctx, deploymentName, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("回滚 Deployment 失败: %w", err)
	}

	logs.Info(map[string]interface{}{"deployment": deploymentName}, "成功回滚 Deployment spec")
	return nil
}

// EvictPodsInDeployment 驱逐属于指定 Deployment 的所有 Pod
func (s *EvictionService) EvictPodsInDeployment(ctx context.Context, instanceID uint, namespace, deploymentName string) error {
	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		return fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	// 获取 Deployment，找到其 selector
	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	labelSelector := metav1.FormatLabelSelector(deploy.Spec.Selector)
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("获取 Pod 列表失败: %w", err)
	}

	gracePeriod := int64(30)
	for _, pod := range podList.Items {
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: namespace,
			},
			DeleteOptions: &metav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
			},
		}
		if err := client.PolicyV1().Evictions(namespace).Evict(ctx, eviction); err != nil {
			logs.Info(map[string]interface{}{"pod": pod.Name, "err": err.Error()}, "驱逐 Pod 失败，跳过")
		} else {
			logs.Info(map[string]interface{}{"pod": pod.Name}, "成功驱逐 Pod")
		}
	}
	return nil
}

// WaitPodsOnChaosNode 等待属于 Deployment 的 Pod 在指定节点上变为 Running（超时 120 秒）
func (s *EvictionService) WaitPodsOnChaosNode(ctx context.Context, instanceID uint, namespace, deploymentName, nodeName string) error {
	client, exists := configs.GetK8sClient(instanceID)
	if !exists {
		return fmt.Errorf("K8s client not initialized for instance %d", instanceID)
	}

	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	labelSelector := metav1.FormatLabelSelector(deploy.Spec.Selector)
	timeout := time.After(120 * time.Second)
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("等待 Pod 迁移到演练节点超时（120s）")
		case <-tick.C:
			podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				LabelSelector: labelSelector,
				FieldSelector: "spec.nodeName=" + nodeName,
			})
			if err != nil {
				continue
			}
			readyCount := 0
			for _, pod := range podList.Items {
				if pod.Status.Phase == corev1.PodRunning {
					allReady := true
					for _, cs := range pod.Status.ContainerStatuses {
						if !cs.Ready {
							allReady = false
							break
						}
					}
					if allReady {
						readyCount++
					}
				}
			}
			expectedReplicas := int32(1)
			if deploy.Spec.Replicas != nil {
				expectedReplicas = *deploy.Spec.Replicas
			}
			if int32(readyCount) >= expectedReplicas {
				logs.Info(map[string]interface{}{
					"deployment": deploymentName,
					"node":       nodeName,
					"readyPods":  readyCount,
				}, "Pod 已在演练节点就绪")
				return nil
			}
		}
	}
}

// PrepareEviction 完整执行演练节点准备流程（Taint → Patch → Evict → Wait）
// 返回原始 spec JSON 供清理时使用
func (s *EvictionService) PrepareEviction(ctx context.Context, instanceID uint, req *reqK8s.PrepareEvictionRequest) (string, error) {
	// Step 1: 给演练节点打 Taint
	if err := s.TaintChaosNode(ctx, instanceID, req.NodeName); err != nil {
		return "", fmt.Errorf("打 Taint 失败: %w", err)
	}

	// Step 2: Patch Deployment（Toleration + NodeAffinity）
	origSpecJSON, err := s.PatchDeploymentForChaos(ctx, instanceID, req.Namespace, req.DeploymentName, req.NodeName)
	if err != nil {
		// 回滚 Taint
		_ = s.UnTaintChaosNode(ctx, instanceID, req.NodeName)
		return "", fmt.Errorf("Patch Deployment 失败: %w", err)
	}

	// Step 3: Evict Pod
	if err := s.EvictPodsInDeployment(ctx, instanceID, req.Namespace, req.DeploymentName); err != nil {
		return origSpecJSON, fmt.Errorf("驱逐 Pod 失败: %w", err)
	}

	// Step 4: 等待 Pod 在演练节点就绪
	if err := s.WaitPodsOnChaosNode(ctx, instanceID, req.Namespace, req.DeploymentName, req.NodeName); err != nil {
		return origSpecJSON, fmt.Errorf("等待 Pod 迁移超时: %w", err)
	}

	return origSpecJSON, nil
}

// CleanupEviction 清理演练环境（回滚 Deployment → Evict → 去 Taint）
func (s *EvictionService) CleanupEviction(ctx context.Context, instanceID uint, req *reqK8s.CleanupEvictionRequest, origSpecJSON string) error {
	// Step 1: 回滚 Deployment spec
	if err := s.RestoreDeployment(ctx, instanceID, req.Namespace, req.DeploymentName, origSpecJSON); err != nil {
		logs.Info(map[string]interface{}{"err": err.Error()}, "回滚 Deployment 失败，继续清理")
	}

	// Step 2: 再次 Evict，让 Pod 回正常节点
	if err := s.EvictPodsInDeployment(ctx, instanceID, req.Namespace, req.DeploymentName); err != nil {
		logs.Info(map[string]interface{}{"err": err.Error()}, "re-Evict Pod 失败，继续清理")
	}

	// Step 3: 去掉节点 Taint
	if err := s.UnTaintChaosNode(ctx, instanceID, req.NodeName); err != nil {
		return fmt.Errorf("去 Taint 失败: %w", err)
	}

	logs.Info(map[string]interface{}{
		"node":       req.NodeName,
		"deployment": req.DeploymentName,
	}, "演练环境清理完成")
	return nil
}

// buildPodDisruptionBudget 辅助：创建临时 PDB 避免驱逐失败（可选）
func buildPodDisruptionBudget(name, namespace string) *policyv1.PodDisruptionBudget {
	minAvail := intstr.FromInt(0)
	return &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "chaos-pdb-" + name,
			Namespace: namespace,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvail,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
		},
	}
}
