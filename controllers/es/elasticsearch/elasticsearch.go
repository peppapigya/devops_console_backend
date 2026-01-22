package elasticsearch

import (
	"context"
	"devops-console-backend/config"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/logs"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
)

// ClusterInfo 集群基本信息结构体
type ClusterInfo struct {
	ClusterName string      `json:"cluster_name"`
	ClusterUUID string      `json:"cluster_uuid"`
	Name        string      `json:"name"`
	VersionInfo interface{} `json:"version_info"`
	Version     string      `json:"version"`
}

// HealthInfo 集群健康状态结构体
type HealthInfo struct {
	Status                      string  `json:"status"`
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
	DocCount                    int64   `json:"doc_count"`
}

// ClusterStatus 集群状态结构体
type ClusterStatus struct {
	ClusterName                 string  `json:"cluster_name"`
	ClusterUUID                 string  `json:"cluster_uuid"`
	VersionInfo                 string  `json:"version_info"`
	Status                      string  `json:"status"`
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	NumberOfDataNodes           int     `json:"number_of_data_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
}

// NodeInfo 节点信息结构体
type NodeInfo struct {
	Name    string      `json:"name"`
	ID      string      `json:"id"`
	IP      string      `json:"ip"`
	Host    string      `json:"host"`
	Version string      `json:"version"`
	Role    string      `json:"role"`
	OS      interface{} `json:"os"`
	JVM     interface{} `json:"jvm"`
	Network interface{} `json:"network"`
}

// ClusterStats 集群统计信息
type ClusterStats struct {
	Nodes   NodeStats  `json:"nodes"`
	Indices IndexStats `json:"indices"`
	Shards  ShardStats `json:"shards"`
}

// NodeStats 节点统计
type NodeStats struct {
	Count NodeCount `json:"count"`
}

// NodeCount 节点计数
type NodeCount struct {
	Total        int `json:"total"`
	Master       int `json:"master"`
	Data         int `json:"data"`
	Ingest       int `json:"ingest"`
	Coordinating int `json:"coordinating"`
}

// IndexStats 索引统计
type IndexStats struct {
	Count int64 `json:"count"`
	Docs  Docs  `json:"docs"`
	Store Store `json:"store"`
}

// Docs 文档统计
type Docs struct {
	Count   int64 `json:"count"`
	Deleted int64 `json:"deleted"`
}

// Store 存储统计
type Store struct {
	SizeInBytes int64  `json:"size_in_bytes"`
	Size        string `json:"size"`
}

// ShardStats 分片统计
type ShardStats struct {
	Total     int `json:"total"`
	Primaries int `json:"primaries"`
}

// GetClusterStatus 获取集群状态信息
func GetClusterStatus(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	// 获取实例ID参数
	instanceIDStr := c.Query("instance_id")
	if instanceIDStr == "" {
		helper.BadRequest("缺少instance_id参数")
		return
	}

	instanceID64, err := strconv.ParseInt(instanceIDStr, 10, 64)
	instanceID := uint(instanceID64)
	if err != nil {
		helper.BadRequest("无效的instance_id参数")
		return
	}

	// 获取ES客户端
	client, exists := config.GetEsClient(instanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 获取集群状态
	status, err := getClusterStatusFromES(client)
	if err != nil {
		helper.DatabaseError("获取集群状态失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{
		"instance_id":  instanceID,
		"cluster_name": status.ClusterName,
		"status":       status.Status,
	}, "获取集群状态成功")
	helper.SuccessWithData("获取集群状态成功", "cluster_status", status)
}

// GetClusterHealthHandler 获取集群健康状态
func GetClusterHealthHandler(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	// 获取实例ID参数
	instanceIDStr := c.Query("instance_id")
	if instanceIDStr == "" {
		helper.BadRequest("缺少instance_id参数")
		return
	}

	instanceID64, err := strconv.ParseInt(instanceIDStr, 10, 64)
	instanceID := uint(instanceID64)
	if err != nil {
		helper.BadRequest("无效的instance_id参数")
		return
	}

	// 获取ES客户端
	client, exists := config.GetEsClient(instanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 获取集群健康状态
	health, err := getClusterHealthFromES(client)
	if err != nil {
		helper.DatabaseError("获取集群健康状态失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{
		"instance_id": instanceID,
		"status":      health.Status,
	}, "获取集群健康状态成功")
	helper.SuccessWithData("获取集群健康状态成功", "health_info", health)
}

// GetClusterInfoHandler 获取集群基本信息
func GetClusterInfoHandler(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	// 获取实例ID参数
	instanceIDStr := c.Query("instance_id")
	if instanceIDStr == "" {
		helper.BadRequest("缺少instance_id参数")
		return
	}

	instanceID64, err := strconv.ParseInt(instanceIDStr, 10, 64)
	instanceID := uint(instanceID64)
	if err != nil {
		helper.BadRequest("无效的instance_id参数")
		return
	}

	// 获取ES客户端
	client, exists := config.GetEsClient(instanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 获取集群基本信息
	info, err := getClusterInfoFromES(client)
	if err != nil {
		helper.DatabaseError("获取集群信息失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{
		"instance_id":  instanceID,
		"cluster_name": info.ClusterName,
	}, "获取集群信息成功")
	helper.SuccessWithData("获取集群信息成功", "cluster_info", info)
}

// GetFullClusterInfo 获取完整的集群信息
func GetFullClusterInfo(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	// 获取实例ID参数
	instanceIDStr := c.Query("instance_id")
	if instanceIDStr == "" {
		helper.BadRequest("缺少instance_id参数")
		return
	}

	instanceID64, err := strconv.ParseInt(instanceIDStr, 10, 64)
	instanceID := uint(instanceID64)
	if err != nil {
		helper.BadRequest("无效的instance_id参数")
		return
	}

	// 获取ES客户端
	client, exists := config.GetEsClient(instanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 并行获取集群信息、健康状态、节点信息和统计信息
	clusterInfoCh := make(chan *ClusterInfo, 1)
	healthInfoCh := make(chan *HealthInfo, 1)
	nodeInfoCh := make(chan []NodeInfo, 1)
	statsInfoCh := make(chan *ClusterStats, 1)
	errCh := make(chan error, 4)

	// 获取集群基本信息
	go func() {
		info, err := getClusterInfoFromES(client)
		if err != nil {
			errCh <- err
			return
		}
		clusterInfoCh <- info
	}()

	// 获取集群健康状态
	go func() {
		health, err := getClusterHealthFromES(client)
		if err != nil {
			errCh <- err
			return
		}
		healthInfoCh <- health
	}()

	// 获取节点信息
	go func() {
		nodes, err := getNodeInfoFromES(client)
		if err != nil {
			errCh <- err
			return
		}
		nodeInfoCh <- nodes
	}()

	// 获取集群统计信息
	go func() {
		stats, err := getClusterStatsFromES(client)
		if err != nil {
			errCh <- err
			return
		}
		statsInfoCh <- stats
	}()

	// 收集结果
	var clusterInfo *ClusterInfo
	var healthInfo *HealthInfo
	var nodeInfo []NodeInfo
	var statsInfo *ClusterStats

	for i := 0; i < 4; i++ {
		select {
		case info := <-clusterInfoCh:
			clusterInfo = info
		case health := <-healthInfoCh:
			healthInfo = health
		case nodes := <-nodeInfoCh:
			nodeInfo = nodes
		case stats := <-statsInfoCh:
			statsInfo = stats
		case err := <-errCh:
			helper.DatabaseError("获取集群信息失败: " + err.Error())
			return
		case <-time.After(30 * time.Second):
			helper.DatabaseError("获取集群信息超时")
			return
		}
	}

	// 组合返回完整的集群信息
	result := gin.H{
		"cluster_info": clusterInfo,
		"health_info":  healthInfo,
		"node_info":    nodeInfo,
		"stats_info":   statsInfo,
	}

	logs.Info(map[string]interface{}{
		"instance_id":  instanceID,
		"cluster_name": clusterInfo.ClusterName,
		"node_count":   len(nodeInfo),
	}, "获取完整集群信息成功")
	helper.SuccessWithData("获取完整集群信息成功", "cluster_info", result)
}

// getClusterStatusFromES 从ES获取集群状态
func getClusterStatusFromES(client *elasticsearch.Client) (*ClusterStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 并行获取集群信息和健康状态
	clusterInfoCh := make(chan *ClusterInfo, 1)
	healthInfoCh := make(chan *HealthInfo, 1)
	errCh := make(chan error, 2)

	// 获取集群基本信息
	go func() {
		info, err := getClusterInfoFromES(client)
		if err != nil {
			errCh <- err
			return
		}
		clusterInfoCh <- info
	}()

	// 获取集群健康状态
	go func() {
		health, err := getClusterHealthFromES(client)
		if err != nil {
			errCh <- err
			return
		}
		healthInfoCh <- health
	}()

	// 收集结果
	var clusterInfo *ClusterInfo
	var healthInfo *HealthInfo

	for i := 0; i < 2; i++ {
		select {
		case info := <-clusterInfoCh:
			clusterInfo = info
		case health := <-healthInfoCh:
			healthInfo = health
		case err := <-errCh:
			return nil, err
		case <-ctx.Done():
			return nil, fmt.Errorf("获取集群状态超时")
		}
	}

	// 组合返回集群状态
	result := &ClusterStatus{
		ClusterName:                 clusterInfo.ClusterName,
		ClusterUUID:                 clusterInfo.ClusterUUID,
		VersionInfo:                 clusterInfo.Version,
		Status:                      healthInfo.Status,
		TimedOut:                    healthInfo.TimedOut,
		NumberOfNodes:               healthInfo.NumberOfNodes,
		NumberOfDataNodes:           healthInfo.NumberOfDataNodes,
		ActivePrimaryShards:         healthInfo.ActivePrimaryShards,
		ActiveShards:                healthInfo.ActiveShards,
		RelocatingShards:            healthInfo.RelocatingShards,
		InitializingShards:          healthInfo.InitializingShards,
		UnassignedShards:            healthInfo.UnassignedShards,
		DelayedUnassignedShards:     healthInfo.DelayedUnassignedShards,
		NumberOfPendingTasks:        healthInfo.NumberOfPendingTasks,
		NumberOfInFlightFetch:       healthInfo.NumberOfInFlightFetch,
		TaskMaxWaitingInQueueMillis: healthInfo.TaskMaxWaitingInQueueMillis,
		ActiveShardsPercentAsNumber: healthInfo.ActiveShardsPercentAsNumber,
	}

	return result, nil
}

// getClusterInfoFromES 从ES获取集群基本信息
func getClusterInfoFromES(client *elasticsearch.Client) (*ClusterInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := esapi.InfoRequest{}
	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("执行集群信息查询失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES返回错误: %s", res.Status())
	}

	// 解析响应
	var clusterInfo map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&clusterInfo); err != nil {
		return nil, fmt.Errorf("解析集群信息失败: %v", err)
	}

	// 构建返回结构体
	result := &ClusterInfo{
		ClusterName: fmt.Sprintf("%v", clusterInfo["cluster_name"]),
		ClusterUUID: fmt.Sprintf("%v", clusterInfo["cluster_uuid"]),
		Name:        fmt.Sprintf("%v", clusterInfo["name"]),
		VersionInfo: clusterInfo["version"],
	}

	// 处理版本信息
	if version, exists := clusterInfo["version"]; exists {
		if versionMap, ok := version.(map[string]interface{}); ok {
			versionNumber := versionMap["number"]
			buildHash := versionMap["build_hash"]
			buildDate := versionMap["build_date"]
			result.Version = fmt.Sprintf("Elasticsearch %v (构建: %v, 构建日期: %v)",
				versionNumber, buildHash, buildDate)
		}
	}

	return result, nil
}

// getClusterHealthFromES 从ES获取集群健康状态
func getClusterHealthFromES(client *elasticsearch.Client) (*HealthInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := esapi.ClusterHealthRequest{}
	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("执行健康状态查询失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES返回错误: %s", res.Status())
	}

	// 解析响应
	var healthInfo map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&healthInfo); err != nil {
		return nil, fmt.Errorf("解析健康状态失败: %v", err)
	}

	// 构建返回结构体
	result := &HealthInfo{
		Status:                      fmt.Sprintf("%v", healthInfo["status"]),
		TimedOut:                    healthInfo["timed_out"].(bool),
		NumberOfNodes:               int(healthInfo["number_of_nodes"].(float64)),
		NumberOfDataNodes:           int(healthInfo["number_of_data_nodes"].(float64)),
		ActivePrimaryShards:         int(healthInfo["active_primary_shards"].(float64)),
		ActiveShards:                int(healthInfo["active_shards"].(float64)),
		RelocatingShards:            int(healthInfo["relocating_shards"].(float64)),
		InitializingShards:          int(healthInfo["initializing_shards"].(float64)),
		UnassignedShards:            int(healthInfo["unassigned_shards"].(float64)),
		DelayedUnassignedShards:     int(healthInfo["delayed_unassigned_shards"].(float64)),
		NumberOfPendingTasks:        int(healthInfo["number_of_pending_tasks"].(float64)),
		NumberOfInFlightFetch:       int(healthInfo["number_of_in_flight_fetch"].(float64)),
		TaskMaxWaitingInQueueMillis: int(healthInfo["task_max_waiting_in_queue_millis"].(float64)),
		ActiveShardsPercentAsNumber: healthInfo["active_shards_percent_as_number"].(float64),
	}

	// 尝试获取文档数量（如果存在）
	if docCount, ok := healthInfo["doc_count"]; ok {
		if docCountFloat, ok := docCount.(float64); ok {
			result.DocCount = int64(docCountFloat)
		}
	}

	return result, nil
}

// getNodeInfoFromES 从ES获取节点信息
func getNodeInfoFromES(client *elasticsearch.Client) ([]NodeInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := esapi.CatNodesRequest{
		Format: "json",
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("执行节点信息查询失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES返回错误: %s", res.Status())
	}

	// 解析响应
	var nodeInfos []NodeInfo
	if err := json.NewDecoder(res.Body).Decode(&nodeInfos); err != nil {
		return nil, fmt.Errorf("解析节点信息失败: %v", err)
	}

	return nodeInfos, nil
}

// getClusterStatsFromES 从ES获取集群统计信息
func getClusterStatsFromES(client *elasticsearch.Client) (*ClusterStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := esapi.ClusterStatsRequest{}
	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("执行集群统计查询失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES返回错误: %s", res.Status())
	}

	// 解析响应
	var statsResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&statsResponse); err != nil {
		return nil, fmt.Errorf("解析集群统计失败: %v", err)
	}

	// 构建统计信息
	stats := &ClusterStats{}

	// 解析节点统计
	if nodesData, ok := statsResponse["nodes"].(map[string]interface{}); ok {
		if count, ok := nodesData["count"].(map[string]interface{}); ok {
			stats.Nodes.Count = parseNodeCount(count)
		}
	}

	// 解析索引统计
	if indicesData, ok := statsResponse["indices"].(map[string]interface{}); ok {
		stats.Indices = parseIndexStats(indicesData)
	}

	// 解析分片统计
	if shardsData, ok := statsResponse["shards"].(map[string]interface{}); ok {
		stats.Shards = parseShardStats(shardsData)
	}

	return stats, nil
}

// parseNodeCount 解析节点计数
func parseNodeCount(data map[string]interface{}) NodeCount {
	result := NodeCount{}

	if total, ok := data["total"].(float64); ok {
		result.Total = int(total)
	}
	if master, ok := data["master"].(map[string]interface{}); ok {
		if count, ok := master["count"].(float64); ok {
			result.Master = int(count)
		}
	}
	if dataNodes, ok := data["data"].(map[string]interface{}); ok {
		if count, ok := dataNodes["count"].(float64); ok {
			result.Data = int(count)
		}
	}
	if ingestNodes, ok := data["ingest"].(map[string]interface{}); ok {
		if count, ok := ingestNodes["count"].(float64); ok {
			result.Ingest = int(count)
		}
	}
	if coordinatingNodes, ok := data["coordinating_only"].(map[string]interface{}); ok {
		if count, ok := coordinatingNodes["count"].(float64); ok {
			result.Coordinating = int(count)
		}
	}

	return result
}

// parseIndexStats 解析索引统计
func parseIndexStats(data map[string]interface{}) IndexStats {
	result := IndexStats{}

	if count, ok := data["count"].(float64); ok {
		result.Count = int64(count)
	}

	if docs, ok := data["docs"].(map[string]interface{}); ok {
		if docCount, ok := docs["count"].(float64); ok {
			result.Docs.Count = int64(docCount)
		}
		if deletedCount, ok := docs["deleted"].(float64); ok {
			result.Docs.Deleted = int64(deletedCount)
		}
	}

	if store, ok := data["store"].(map[string]interface{}); ok {
		if size, ok := store["size_in_bytes"].(float64); ok {
			result.Store.SizeInBytes = int64(size)
			result.Store.Size = formatBytes(int64(size))
		}
	}

	return result
}

// parseShardStats 解析分片统计
func parseShardStats(data map[string]interface{}) ShardStats {
	result := ShardStats{}

	if total, ok := data["total"].(float64); ok {
		result.Total = int(total)
	}
	if primaries, ok := data["primaries"].(float64); ok {
		result.Primaries = int(primaries)
	}

	return result
}

// formatBytes 格式化字节数
func formatBytes(byteSize int64) string {
	const unit = 1024
	if byteSize < unit {
		return fmt.Sprintf("%d B", byteSize)
	}
	div, exp := int64(unit), 0
	for n := byteSize / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(byteSize)/float64(div), "KMGTPE"[exp])
}
