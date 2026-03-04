package probe

import (
	"context"
	"devops-console-backend/internal/dal"
	"devops-console-backend/pkg/utils/logs"
	"sync"
)

type InstanceProbeTYpe string

// 这个要和数据库实例 type 的值相同
const (
	PrometheusInstanceProbeType    InstanceProbeTYpe = "Prometheus"
	KubernetesInstanceProbeType    InstanceProbeTYpe = "Kubernetes"
	InstanceProbeTypeElasticsearch InstanceProbeTYpe = "Elasticsearch"
)

// Prober 是所有探测器的统一接口
type Prober interface {
	// SupportType 返回探测器支持的 Instance_Type_Name (如 "kubernetes", "ssh", "prometheus")
	SupportType() InstanceProbeTYpe

	// Probe 执行一次具体的探测逻辑，返回状态 (StatusOnline 或 StatusOffline)
	Probe(ctx context.Context, instance dal.Instance) string
}

var (
	probers = make(map[string]Prober)
	mu      sync.RWMutex
)

// RegisterProber 注册一个探测器
func RegisterProber(p Prober) {
	mu.Lock()
	defer mu.Unlock()

	if p == nil {
		logs.Error(nil, "Prober 不能为空")
		return
	}
	probers[string(p.SupportType())] = p
	logs.Info(map[string]interface{}{"type": p.SupportType()}, "Prober 注册成功")
}

// GetProber 获取对应类型的探测器
func GetProber(instanceTypeName string) (Prober, bool) {
	mu.RLock()
	defer mu.RUnlock()

	p, ok := probers[instanceTypeName]
	return p, ok
}
