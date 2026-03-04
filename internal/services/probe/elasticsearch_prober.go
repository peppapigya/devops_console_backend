package probe

import (
	"context"
	"devops-console-backend/internal/dal"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils/logs"
	"time"
)

type ElacsticsearchProber struct{}

func (e *ElacsticsearchProber) SupportType() InstanceProbeTYpe {
	return InstanceProbeTypeElasticsearch
}

func (e *ElacsticsearchProber) Probe(ctx context.Context, instance dal.Instance) string {
	if instance.ID == 0 {
		return StatusOffline
	}
	client, exist := configs.GetEsClient(instance.ID)
	if !exist {
		logs.Error(map[string]interface{}{"instance_id": instance.ID}, "ES 集群客户端不存在")
		return StatusOffline
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	health, err := client.Cluster.Health(client.Cluster.Health.WithContext(ctx))
	if err != nil {
		logs.Error(map[string]interface{}{"instanceId": instance.ID, "错误原因：": err.Error()}, "ES 集群状态异常")
		return StatusOffline
	}
	if health.StatusCode != 200 {
		return StatusOffline
	}

	return StatusOnline

}

func init() {
	RegisterProber(&ElacsticsearchProber{})
}
