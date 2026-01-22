package shard

import (
	"bytes"
	"context"
	"devops-console-backend/config"
	"devops-console-backend/pkg/utils"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
)

// ShardInfo 分片信息结构体
type ShardInfo struct {
	Index  string `json:"index"`
	Shard  string `json:"shard"`
	Prirep string `json:"prirep"`
	State  string `json:"state"`
	Docs   string `json:"docs"`
	Store  string `json:"store"`
	IP     string `json:"ip"`
	Node   string `json:"node"`
}

// ShardStats 分片统计信息结构体
type ShardStats struct {
	Indices      map[string]IndexShardStats `json:"indices"`
	TotalShards  int                        `json:"total_shards"`
	ActiveShards int                        `json:"active_shards"`
}

// IndexShardStats 索引分片统计
type IndexShardStats struct {
	Primaries ShardCount `json:"primaries"`
	Total     ShardCount `json:"total"`
}

// ShardCount 分片计数
type ShardCount struct {
	Docs      int64  `json:"docs"`
	Store     string `json:"store"`
	Shards    int    `json:"shards"`
	PriShards int    `json:"pri_shards"`
}

// GetShardInfo 获取分片信息
func GetShardInfo(c *gin.Context) {
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

	// 获取分片信息
	shards, err := getShardInfoFromES(client)
	if err != nil {
		helper.DatabaseError("获取分片信息失败: " + err.Error())
		return
	}

	helper.SuccessWithData("获取分片信息成功", "shards", shards)
}

// GetShardStats 获取分片统计信息
func GetShardStats(c *gin.Context) {
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

	// 获取分片统计信息
	stats, err := getShardStatsFromES(client)
	if err != nil {
		helper.DatabaseError("获取分片统计失败: " + err.Error())
		return
	}

	helper.SuccessWithData("获取分片统计成功", "stats", stats)
}

// AllocateShard 手动分配分片
func AllocateShard(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	var req struct {
		InstanceID uint   `json:"instance_id" binding:"required"`
		Index      string `json:"index" binding:"required"`
		Shard      int    `json:"shard" binding:"required"`
		Node       string `json:"node" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数绑定失败: " + err.Error())
		return
	}

	// 获取ES客户端
	client, exists := config.GetEsClient(req.InstanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 执行分片分配
	err := allocateShardToNode(client, req.Index, req.Shard, req.Node)
	if err != nil {
		helper.DatabaseError("分片分配失败: " + err.Error())
		return
	}

	helper.Success("分片分配成功")
}

// MoveShard 移动分片
func MoveShard(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	var req struct {
		InstanceID uint   `json:"instance_id" binding:"required"`
		Index      string `json:"index" binding:"required"`
		Shard      int    `json:"shard" binding:"required"`
		FromNode   string `json:"from_node" binding:"required"`
		ToNode     string `json:"to_node" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数绑定失败: " + err.Error())
		return
	}

	// 获取ES客户端
	client, exists := config.GetEsClient(req.InstanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 执行分片移动
	err := moveShardBetweenNodes(client, req.Index, req.Shard, req.FromNode, req.ToNode)
	if err != nil {
		helper.DatabaseError("分片移动失败: " + err.Error())
		return
	}

	helper.Success("分片移动成功")
}

// MigrateShardToAnotherInstance 将分片（索引）迁移到另一个实例
func MigrateShardToAnotherInstance(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	var req struct {
		SourceInstanceID uint   `json:"source_instance_id" binding:"required"`
		TargetInstanceID uint   `json:"target_instance_id" binding:"required"`
		Index            string `json:"index" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数绑定失败: " + err.Error())
		return
	}

	// 获取源ES客户端
	sourceClient, exists := config.GetEsClient(req.SourceInstanceID)
	if !exists {
		helper.BadRequest("源实例不存在或未初始化")
		return
	}

	// 获取目标ES客户端
	targetClient, exists := config.GetEsClient(req.TargetInstanceID)
	if !exists {
		helper.BadRequest("目标实例不存在或未初始化")
		return
	}

	// 检查目标实例是否可访问
	if err := checkInstanceHealth(targetClient); err != nil {
		helper.DatabaseError("目标实例不可访问: " + err.Error())
		return
	}

	// 执行跨实例索引迁移
	err := migrateIndexToAnotherInstance(sourceClient, targetClient, req.Index)
	if err != nil {
		helper.DatabaseError("索引迁移失败: " + err.Error())
		return
	}

	helper.Success("索引迁移成功")
}

// getShardInfoFromES 从ES获取分片信息
func getShardInfoFromES(client *elasticsearch.Client) ([]ShardInfo, error) {
	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 执行_cat/shards API
	req := esapi.CatShardsRequest{
		Format: "json",
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("执行分片查询失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES返回错误: %s", res.Status())
	}

	// 解析响应
	var shards []ShardInfo
	if err := json.NewDecoder(res.Body).Decode(&shards); err != nil {
		return nil, fmt.Errorf("解析分片信息失败: %v", err)
	}

	return shards, nil
}

// getShardStatsFromES 从ES获取分片统计信息
func getShardStatsFromES(client *elasticsearch.Client) (*ShardStats, error) {
	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 执行_stats API
	req := esapi.IndicesStatsRequest{}
	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("执行统计查询失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES返回错误: %s", res.Status())
	}

	// 解析响应
	var statsResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&statsResponse); err != nil {
		return nil, fmt.Errorf("解析统计信息失败: %v", err)
	}

	// 构建简化的统计信息
	stats := &ShardStats{
		Indices: make(map[string]IndexShardStats),
	}

	// 提取索引统计信息
	if indices, ok := statsResponse["indices"].(map[string]interface{}); ok {
		for indexName, indexData := range indices {
			if indexMap, ok := indexData.(map[string]interface{}); ok {
				if primaries, ok := indexMap["primaries"].(map[string]interface{}); ok {
					indexStats := IndexShardStats{
						Primaries: parseShardCount(primaries),
					}
					if total, ok := indexMap["total"].(map[string]interface{}); ok {
						indexStats.Total = parseShardCount(total)
					}
					stats.Indices[indexName] = indexStats
				}
			}
		}
	}

	// 提取总体统计信息
	if shards, ok := statsResponse["_shards"].(map[string]interface{}); ok {
		if total, ok := shards["total"].(float64); ok {
			stats.TotalShards = int(total)
		}
		if active, ok := shards["successful"].(float64); ok {
			stats.ActiveShards = int(active)
		}
	}

	return stats, nil
}

// parseShardCount 解析分片计数
func parseShardCount(data map[string]interface{}) ShardCount {
	result := ShardCount{}

	if docs, ok := data["docs"].(map[string]interface{}); ok {
		if docCount, ok := docs["count"].(float64); ok {
			result.Docs = int64(docCount)
		}
	}

	if store, ok := data["store"].(map[string]interface{}); ok {
		if size, ok := store["size_in_bytes"].(float64); ok {
			result.Store = formatBytes(int64(size))
		}
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

// allocateShardToNode 分配分片到指定节点
func allocateShardToNode(client *elasticsearch.Client, index string, shard int, node string) error {
	// 构建分配命令
	command := map[string]interface{}{
		"commands": []map[string]interface{}{
			{
				"allocate_primary": map[string]interface{}{
					"index": index,
					"shard": shard,
					"node":  node,
				},
			},
		},
	}

	// 序列化命令
	cmdBytes, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("序列化分配命令失败: %v", err)
	}

	// 执行分配命令
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := esapi.ClusterRerouteRequest{
		Body: bytes.NewReader(cmdBytes),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("执行分配命令失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES分配失败: %s", res.Status())
	}

	return nil
}

// moveShardBetweenNodes 在节点间移动分片
func moveShardBetweenNodes(client *elasticsearch.Client, index string, shard int, fromNode, toNode string) error {
	// 构建移动命令
	command := map[string]interface{}{
		"commands": []map[string]interface{}{
			{
				"move": map[string]interface{}{
					"index":     index,
					"shard":     shard,
					"from_node": fromNode,
					"to_node":   toNode,
				},
			},
		},
	}

	// 序列化命令
	cmdBytes, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("序列化移动命令失败: %v", err)
	}

	// 执行移动命令
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := esapi.ClusterRerouteRequest{
		Body: bytes.NewReader(cmdBytes),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("执行移动命令失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES移动失败: %s", res.Status())
	}

	return nil
}

// checkInstanceHealth 检查ES实例健康状态
func checkInstanceHealth(client *elasticsearch.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := esapi.InfoRequest{}
	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("获取实例信息失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES实例返回错误: %s", res.Status())
	}

	return nil
}

// migrateIndexToAnotherInstance 将索引从一个实例迁移到另一个实例
func migrateIndexToAnotherInstance(sourceClient, targetClient *elasticsearch.Client, index string) error {
	// 1. 获取源索引的映射和设置
	mapping, settings, err := getIndexMetadata(sourceClient, index)
	if err != nil {
		return fmt.Errorf("获取源索引元数据失败: %v", err)
	}

	// 2. 在目标实例上创建索引（使用相同的映射和设置）
	targetIndex := index // 可以选择使用相同名称或其他命名规则
	err = createIndexOnTarget(targetClient, targetIndex, mapping, settings)
	if err != nil {
		return fmt.Errorf("在目标实例上创建索引失败: %v", err)
	}

	// 3. 执行reindex操作
	err = performReindex(sourceClient, targetClient, index, targetIndex)
	if err != nil {
		return fmt.Errorf("执行reindex操作失败: %v", err)
	}

	return nil
}

// getIndexMetadata 获取索引的映射和设置
func getIndexMetadata(client *elasticsearch.Client, index string) (map[string]interface{}, map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取索引映射
	req := esapi.IndicesGetMappingRequest{Index: []string{index}}
	mappingRes, err := req.Do(ctx, client)
	if err != nil {
		return nil, nil, fmt.Errorf("获取映射失败: %v", err)
	}
	defer mappingRes.Body.Close()

	if mappingRes.IsError() {
		return nil, nil, fmt.Errorf("获取映射返回错误: %s", mappingRes.Status())
	}

	var mapping map[string]interface{}
	if err := json.NewDecoder(mappingRes.Body).Decode(&mapping); err != nil {
		return nil, nil, fmt.Errorf("解析映射失败: %v", err)
	}

	// 获取索引设置
	settingsReq := esapi.IndicesGetSettingsRequest{Index: []string{index}}
	settingsRes, err := settingsReq.Do(ctx, client)
	if err != nil {
		return nil, nil, fmt.Errorf("获取设置失败: %v", err)
	}
	defer settingsRes.Body.Close()

	if settingsRes.IsError() {
		return nil, nil, fmt.Errorf("获取设置返回错误: %s", settingsRes.Status())
	}

	var settings map[string]interface{}
	if err := json.NewDecoder(settingsRes.Body).Decode(&settings); err != nil {
		return nil, nil, fmt.Errorf("解析设置失败: %v", err)
	}

	return mapping, settings, nil
}

// createIndexOnTarget 在目标实例上创建索引
func createIndexOnTarget(client *elasticsearch.Client, index string, mapping, settings map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 构建创建索引的请求体
	body := make(map[string]interface{})

	// 添加映射
	if indexMapping, ok := mapping[index]; ok {
		if mappings, exists := indexMapping.(map[string]interface{})["mappings"]; exists {
			body["mappings"] = mappings
		}
	}

	// 添加设置
	if indexSettings, ok := settings[index]; ok {
		if indexSettingsMap, exists := indexSettings.(map[string]interface{})["settings"]; exists {
			body["settings"] = indexSettingsMap
		}
	}

	// 序列化请求体
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 创建索引
	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  bytes.NewReader(bodyBytes),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("创建索引返回错误: %s", res.Status())
	}

	return nil
}

// performReindex 执行reindex操作
func performReindex(sourceClient, targetClient *elasticsearch.Client, sourceIndex, targetIndex string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // 设置较长的超时时间
	defer cancel()

	// 构建reindex请求体
	reindexBody := map[string]interface{}{
		"source": map[string]interface{}{
			"index": sourceIndex,
		},
		"dest": map[string]interface{}{
			"index": targetIndex,
		},
	}

	// 序列化reindex请求
	bodyBytes, err := json.Marshal(reindexBody)
	if err != nil {
		return fmt.Errorf("序列化reindex请求体失败: %v", err)
	}

	// 使用Reindex API进行迁移
	req := esapi.ReindexRequest{
		Body: bytes.NewReader(bodyBytes),
	}

	res, err := req.Do(ctx, targetClient)
	if err != nil {
		return fmt.Errorf("执行reindex失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("reindex操作返回错误: %s", res.Status())
	}

	// 读取响应以确认操作完成
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("解析reindex响应失败: %v", err)
	}

	return nil
}

// ClusterReroute 执行集群重路由操作，支持多种操作类型（移动、分配、取消分配等）
func ClusterReroute(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	var req struct {
		InstanceID uint                     `json:"instance_id" binding:"required"`
		Commands   []map[string]interface{} `json:"commands" binding:"required"`
		DryRun     bool                     `json:"dry_run"` // 是否只做预览
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数绑定失败: " + err.Error())
		return
	}

	// 获取ES客户端
	client, exists := config.GetEsClient(req.InstanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 构建重路由命令
	command := map[string]interface{}{
		"commands": req.Commands,
	}

	// 如果是dry run模式，添加dry_run参数
	params := make(map[string]string)
	if req.DryRun {
		params["dry_run"] = "true"
	}

	// 执行重路由操作
	err := executeClusterReroute(client, command, params)
	if err != nil {
		helper.DatabaseError("集群重路由操作失败: " + err.Error())
		return
	}

	if req.DryRun {
		helper.Success("集群重路由预览成功")
	} else {
		helper.Success("集群重路由操作成功")
	}
}

// executeClusterReroute 执行集群重路由
func executeClusterReroute(client *elasticsearch.Client, command map[string]interface{}, params map[string]string) error {
	// 序列化命令
	cmdBytes, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("序列化重路由命令失败: %v", err)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 构建请求参数
	req := esapi.ClusterRerouteRequest{
		Body: bytes.NewReader(cmdBytes),
	}

	// 添加额外参数（如dry_run）
	if params != nil {
		if dryRunStr, ok := params["dry_run"]; ok {
			dryRun := dryRunStr == "true" || dryRunStr == "1"
			req.DryRun = &dryRun
		}
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("执行重路由命令失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES重路由操作失败: %s", res.Status())
	}

	return nil
}
