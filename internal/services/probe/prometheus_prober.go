package probe

import (
	"context"
	"devops-console-backend/internal/dal"
	"devops-console-backend/pkg/utils/logs"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// PrometheusProber Prometheus 实例探测器
type PrometheusProber struct{}

func (p *PrometheusProber) SupportType() InstanceProbeTYpe {
	return PrometheusInstanceProbeType
}

func (p *PrometheusProber) Probe(ctx context.Context, instance dal.Instance) string {
	logs.Info(map[string]interface{}{"instance": instance.Name}, "Prometheus 探测执行")
	address := "http://" + instance.Address
	if instance.HttpsEnabled {
		address = "https://" + instance.Address
	}
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		logs.Error(map[string]interface{}{"instance": instance.Name, "reason": err.Error()}, "Prometheus 探测失败")
		return StatusOffline
	}
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	runtimeInfo, err := v1api.Runtimeinfo(ctx)
	if err != nil {
		logs.Error(map[string]interface{}{"instance": instance.Name, "reason": err.Error()}, "Prometheus 探测失败")
		return StatusOffline
	}
	logs.Info(map[string]interface{}{"instance": instance.Name, "runtimeInfo": runtimeInfo}, "Prometheus 探测成功")
	return StatusOnline
}

func init() {
	RegisterProber(&PrometheusProber{})
}
