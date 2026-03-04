package probe

import (
	"context"
	"devops-console-backend/internal/dal"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils/logs"
	"time"
)

// K8sProber K8s集群探测器
type K8sProber struct{}

// SupportType 实现接口
func (k *K8sProber) SupportType() InstanceProbeTYpe {
	return KubernetesInstanceProbeType
}

// Probe 实现接口：实际的探测逻辑
func (k *K8sProber) Probe(ctx context.Context, instance dal.Instance) string {
	logs.Info(map[string]interface{}{"instance_id": instance.ID}, "开始探测 K8s 集群")
	client, exists := configs.GetK8sClient(instance.ID)
	if !exists {
		logs.Error(map[string]interface{}{"instance_id": instance.ID}, "K8s 集群客户端不存在")
		return StatusOffline
	}

	// 为单次探测加上独立的短超时限制
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 探测 K8s API Server
	res := client.Discovery().RESTClient().Get().AbsPath("/version").Do(probeCtx)
	if res.Error() != nil {
		logs.Error(map[string]interface{}{"instance_id": instance.ID, "reason": res.Error().Error()}, "K8s API Server 响应异常")
		return StatusOffline
	}
	logs.Info(map[string]interface{}{"instance_id": instance.ID}, "K8s API Server 响应正常")
	return StatusOnline
}

func init() {
	RegisterProber(&K8sProber{})
}
