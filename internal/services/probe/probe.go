package probe

import (
	"context"
	"devops-console-backend/internal/dal"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils/logs"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Status 定义探测状态枚举
const (
	StatusOnline  = "online"  // 在线
	StatusOffline = "offline" // 不在线
)

// ProbeResult 探测结果结构
type ProbeResult struct {
	InstanceID uint
	Status     string
}

// isProbing 防止 Cron 重叠执行
var (
	isProbing     bool
	isProbingLock sync.Mutex
)

// StartInstanceStatusProbe 启动全局探测定时任务
func StartInstanceStatusProbe() {
	// 创建支持秒级别的 cron 定时器
	c := cron.New(cron.WithSeconds())

	// 每隔 30 秒执行一次全量探测任务
	_, err := c.AddFunc("0 0 */1 * * *", func() {
		isProbingLock.Lock()
		if isProbing {
			isProbingLock.Unlock()
			logs.Warning(nil, "[Probe Task] 上一次探测未结束，跳过本次执行")
			return
		}
		isProbing = true
		isProbingLock.Unlock()

		defer func() {
			isProbingLock.Lock()
			isProbing = false
			isProbingLock.Unlock()
		}()

		logs.Info(nil, "[Probe Task] 开始执行实例存活探测任务...")

		var taskWg sync.WaitGroup
		taskWg.Add(2)

		go func() {
			defer taskWg.Done()
			executeProbe()
		}()

		go func() {
			defer taskWg.Done()
			executeHostProbe()
		}()

		taskWg.Wait()
		logs.Info(nil, "[Probe Task] 本轮探测任务全部执行完成")
	})

	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "启动探测定时任务失败")
		return
	}

	c.Start()
	logs.Info(nil, "[Probe Task] 定时任务初始化成功!")
}

func executeProbe() {
	// 1. 获取所有的 instance
	// 为了简单，直接用 DB 查所有实例
	var instances []dal.Instance
	if err := configs.GORMDB.Find(&instances).Error; err != nil {
		logs.Error(map[string]interface{}{"err": err.Error()}, "[Probe Task] 查询实例列表失败")
		return
	}

	instancesCount := len(instances)
	logs.Info(map[string]interface{}{"count": instancesCount}, "[Probe Task] 获取实例列表成功")
	if instancesCount == 0 {
		return
	}

	// 2. 环境设定
	var wg sync.WaitGroup
	resultCh := make(chan ProbeResult, instancesCount)

	// 最大并发数
	maxWorkers := 50
	limitCh := make(chan struct{}, maxWorkers)

	// 3. 并发探测
	for _, inst := range instances {
		wg.Add(1)
		limitCh <- struct{}{}

		go func(instance dal.Instance) {
			defer wg.Done()
			defer func() { <-limitCh }()

			status := probeSingleInstance(instance)

			// 如果状态没有变化则尽量减少 DB UPDATE，我们可以取数据库中的原来 status
			oldStatus := instance.Status

			if oldStatus != status && !(oldStatus == "active" && status == StatusOnline) {
				resultCh <- ProbeResult{
					InstanceID: instance.ID,
					Status:     status,
				}
			}
		}(inst)
	}

	// 4. 等待所有并发探测完成
	wg.Wait()
	close(resultCh)

	// 5. 结果聚合与批量更新数据库
	var onlineIDs []uint
	var offlineIDs []uint

	for res := range resultCh {
		if res.Status == StatusOnline {
			onlineIDs = append(onlineIDs, res.InstanceID)
		} else {
			offlineIDs = append(offlineIDs, res.InstanceID)
		}
	}

	updateStatusInBatch(onlineIDs, StatusOnline)
	updateStatusInBatch(offlineIDs, StatusOffline)

	if len(onlineIDs) > 0 || len(offlineIDs) > 0 {
		logs.Info(map[string]interface{}{
			"online_changed_count":  len(onlineIDs),
			"offline_changed_count": len(offlineIDs),
		}, "[Probe Task] 状态变更更新完毕")
	}
}

// probeSingleInstance 实现对单一目标的探测
func probeSingleInstance(inst dal.Instance) string {
	var instanceType dal.InstanceType
	if err := configs.GORMDB.First(&instanceType, inst.InstanceTypeID).Error; err != nil {
		logs.Warning(map[string]interface{}{"instance_id": inst.ID}, "未找到实例类型，无法探测")
		return StatusOffline
	}

	prober, exists := GetProber(instanceType.TypeName)
	if !exists {
		logs.Warning(map[string]interface{}{
			"instance_id": inst.ID,
			"type":        instanceType.TypeName,
		}, "不支持的探测类型，跳过")
		return StatusOffline
	}

	ctx := context.Background()
	return prober.Probe(ctx, inst)
}

func updateStatusInBatch(ids []uint, status string) {
	if len(ids) == 0 {
		return
	}

	// 批量分块更新 (考虑 SQL in (?) 个数上限，可以做 chunk)
	chunkSize := 100
	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[i:end]

		err := configs.GORMDB.Table("instances").Where("id IN ?", chunk).Update("status", status).Error
		if err != nil {
			logs.Error(map[string]interface{}{"error": err.Error(), "status": status}, "批量更新实例状态失败")
		}
	}
}

func executeHostProbe() {
	var hosts []model.AssetHost
	if err := configs.GORMDB.Find(&hosts).Error; err != nil {
		logs.Error(map[string]interface{}{"err": err.Error()}, "[Probe Task] 查询主机列表失败")
		return
	}

	hostsCount := len(hosts)
	if hostsCount == 0 {
		return
	}

	var wg sync.WaitGroup
	resultCh := make(chan ProbeResult, hostsCount)

	maxWorkers := 50
	limitCh := make(chan struct{}, maxWorkers)

	for _, host := range hosts {
		wg.Add(1)
		limitCh <- struct{}{}

		go func(h model.AssetHost) {
			defer wg.Done()
			defer func() { <-limitCh }()

			status := probeSingleHost(h)

			oldStatus := h.Status
			if oldStatus != status && !(oldStatus == "active" && status == StatusOnline) {
				// 我们重用ProbeResult，但cast InstanceID来记录HostID
				resultCh <- ProbeResult{
					InstanceID: uint(h.ID),
					Status:     status,
				}
			}
		}(host)
	}

	wg.Wait()
	close(resultCh)

	var onlineIDs []uint64
	var offlineIDs []uint64

	for res := range resultCh {
		if res.Status == StatusOnline {
			onlineIDs = append(onlineIDs, uint64(res.InstanceID))
		} else {
			offlineIDs = append(offlineIDs, uint64(res.InstanceID))
		}
	}

	updateHostStatusInBatch(onlineIDs, StatusOnline)
	updateHostStatusInBatch(offlineIDs, StatusOffline)

	if len(onlineIDs) > 0 || len(offlineIDs) > 0 {
		logs.Info(map[string]interface{}{
			"online_changed_count":  len(onlineIDs),
			"offline_changed_count": len(offlineIDs),
		}, "[Probe Task] 主机状态变更更新完毕")
	}
}

func probeSingleHost(host model.AssetHost) string {
	hostPort := fmt.Sprintf("%s:%d", host.IP, host.Port)
	// 短超时 3 秒尝试建立 TCP 连接，判断 SSH 端口存活与否
	conn, err := net.DialTimeout("tcp", hostPort, 3*time.Second)
	if err != nil {
		return StatusOffline
	}
	defer func() { _ = conn.Close() }()
	return StatusOnline
}

func updateHostStatusInBatch(ids []uint64, status string) {
	if len(ids) == 0 {
		return
	}

	chunkSize := 100
	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[i:end]

		err := configs.GORMDB.Table("asset_hosts").Where("id IN ?", chunk).Update("status", status).Error
		if err != nil {
			logs.Error(map[string]interface{}{"error": err.Error(), "status": status}, "批量更新主机状态失败")
		}
	}
}
