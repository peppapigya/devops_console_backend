package backup

import (
	"bytes"
	"context"
	"devops-console-backend/config"
	"devops-console-backend/pkg/utils"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
)

// BackupRequest 备份请求结构体
type BackupRequest struct {
	InstanceID         uint                   `json:"instance_id" binding:"required"`
	Repository         string                 `json:"repository" binding:"required"`
	Snapshot           string                 `json:"snapshot" binding:"required"`
	Indices            []string               `json:"indices"`
	IgnoreUnavailable  bool                   `json:"ignore_unavailable"`
	IncludeGlobalState bool                   `json:"include_global_state"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// RestoreRequest 恢复请求结构体
type RestoreRequest struct {
	InstanceID         uint     `json:"instance_id" binding:"required"`
	Repository         string   `json:"repository" binding:"required"`
	Snapshot           string   `json:"snapshot" binding:"required"`
	Indices            []string `json:"indices"`
	IgnoreUnavailable  bool     `json:"ignore_unavailable"`
	IncludeGlobalState bool     `json:"include_global_state"`
	RenamePattern      string   `json:"rename_pattern"`
	RenameReplacement  string   `json:"rename_replacement"`
}

// RepositoryRequest 仓库请求结构体
type RepositoryRequest struct {
	InstanceID uint                   `json:"instance_id" binding:"required"`
	Type       string                 `json:"type" binding:"required"`
	Settings   map[string]interface{} `json:"settings" binding:"required"`
	Verify     bool                   `json:"verify"`
}

// BackupInfo 备份信息结构体
type BackupInfo struct {
	Snapshot           string                 `json:"snapshot"`
	Repository         string                 `json:"repository"`
	UUID               string                 `json:"uuid"`
	Version            string                 `json:"version"`
	StartTime          time.Time              `json:"start_time"`
	EndTime            time.Time              `json:"end_time"`
	Duration           string                 `json:"duration"`
	Indices            []string               `json:"indices"`
	State              string                 `json:"state"`
	Reason             string                 `json:"reason"`
	Shards             map[string]interface{} `json:"shards"`
	IncludeGlobalState bool                   `json:"include_global_state"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// RepositoryInfo 仓库信息结构体
type RepositoryInfo struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings"`
}

// CreateRepository 创建快照仓库
func CreateRepository(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数绑定失败: " + err.Error())
		return
	}

	// 从请求中提取参数
	instanceIDFloat, ok := req["instance_id"].(float64)
	var instanceID uint
	if ok {
		instanceID = uint(instanceIDFloat)
	}
	if !ok {
		helper.BadRequest("无效的instance_id参数")
		return
	}

	repoType, ok := req["type"].(string)
	if !ok {
		helper.BadRequest("仓库类型不能为空")
		return
	}

	settings, ok := req["settings"].(map[string]interface{})
	if !ok {
		helper.BadRequest("仓库设置不能为空")
		return
	}

	verify, _ := req["verify"].(bool)
	repositoryName, ok := settings["location"].(string)
	if !ok {
		helper.BadRequest("仓库位置不能为空")
		return
	}

	// 从路径中提取仓库名称（如果location是完整路径，取最后一部分作为仓库名）
	parts := strings.Split(repositoryName, "/")
	if len(parts) > 0 {
		repositoryName = parts[len(parts)-1]
	}

	// 获取ES客户端
	client, exists := config.GetEsClient(instanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 创建仓库
	err := createSnapshotRepository(client, repositoryName, repoType, settings, verify)
	if err != nil {
		helper.DatabaseError("创建仓库失败: " + err.Error())
		return
	}

	helper.Success("仓库创建成功")
}

// DeleteRepository 删除快照仓库
func DeleteRepository(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	instanceIDStr := c.Query("instance_id")
	repository := c.Query("repository")

	if instanceIDStr == "" || repository == "" {
		helper.BadRequest("缺少必要参数")
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

	// 删除仓库
	err = deleteSnapshotRepository(client, repository)
	if err != nil {
		helper.DatabaseError("删除仓库失败: " + err.Error())
		return
	}

	helper.Success("仓库删除成功")
}

// ListRepositories 列出所有仓库
func ListRepositories(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

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

	// 获取仓库列表
	repositories, err := getSnapshotRepositories(client)
	if err != nil {
		helper.DatabaseError("获取仓库列表失败: " + err.Error())
		return
	}

	helper.SuccessWithData("获取仓库列表成功", "repositories", repositories)
}

// CreateSnapshot 创建快照
func CreateSnapshot(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	var req BackupRequest
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

	// 创建快照
	err := createSnapshot(client, req.Repository, req.Snapshot, req)
	if err != nil {
		helper.DatabaseError("创建快照失败: " + err.Error())
		return
	}

	helper.Success("快照创建任务已提交")
}

// DeleteSnapshot 删除快照
func DeleteSnapshot(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	instanceIDStr := c.Query("instance_id")
	repository := c.Query("repository")
	snapshot := c.Query("snapshot")

	if instanceIDStr == "" || repository == "" || snapshot == "" {
		helper.BadRequest("缺少必要参数")
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

	// 删除快照
	err = deleteSnapshot(client, repository, snapshot)
	if err != nil {
		helper.DatabaseError("删除快照失败: " + err.Error())
		return
	}

	helper.Success("快照删除成功")
}

// ListSnapshots 列出快照
func ListSnapshots(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	instanceIDStr := c.Query("instance_id")
	repository := c.Query("repository")

	if instanceIDStr == "" || repository == "" {
		helper.BadRequest("缺少必要参数")
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

	// 获取快照列表
	snapshots, err := getSnapshots(client, repository)
	if err != nil {
		helper.DatabaseError("获取快照列表失败: " + err.Error())
		return
	}

	helper.SuccessWithData("获取快照列表成功", "snapshots", snapshots)
}

// GetSnapshotStatus 获取快照状态
func GetSnapshotStatus(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	instanceIDStr := c.Query("instance_id")
	repository := c.Query("repository")
	snapshot := c.Query("snapshot")

	if instanceIDStr == "" || repository == "" || snapshot == "" {
		helper.BadRequest("缺少必要参数")
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

	// 获取快照状态
	status, err := getSnapshotStatus(client, repository, snapshot)
	if err != nil {
		helper.DatabaseError("获取快照状态失败: " + err.Error())
		return
	}

	helper.SuccessWithData("获取快照状态成功", "status", status)
}

// RestoreSnapshot 恢复快照
func RestoreSnapshot(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	var req RestoreRequest
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

	// 恢复快照
	err := restoreSnapshot(client, req.Repository, req.Snapshot, req)
	if err != nil {
		helper.DatabaseError("恢复快照失败: " + err.Error())
		return
	}

	helper.Success("快照恢复任务已提交")
}

// GetRestoreStatus 获取恢复状态
func GetRestoreStatus(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

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

	// 获取恢复状态
	status, err := getRestoreStatus(client)
	if err != nil {
		helper.DatabaseError("获取恢复状态失败: " + err.Error())
		return
	}

	helper.SuccessWithData("获取恢复状态成功", "status", status)
}

// createSnapshotRepository 创建快照仓库
func createSnapshotRepository(client *elasticsearch.Client, repositoryName string, repoType string, settings map[string]interface{}, verify bool) error {
	// 构建仓库配置
	repoConfig := map[string]interface{}{
		"type":     repoType,
		"settings": settings,
	}

	// 序列化配置
	configBytes, err := json.Marshal(repoConfig)
	if err != nil {
		return fmt.Errorf("序列化仓库配置失败: %v", err)
	}

	// 执行创建请求
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := esapi.SnapshotCreateRepositoryRequest{
		Repository: repositoryName, // 使用具体的仓库名称
		Body:       bytes.NewReader(configBytes),
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("执行创建仓库请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// 读取错误响应体以获取更详细的错误信息
		errorBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("ES创建仓库失败: %s, 错误详情: %s", res.Status(), string(errorBody))
	}

	// 如果需要验证
	if verify {
		return verifyRepository(client, repositoryName)
	}

	return nil
}

// deleteSnapshotRepository 删除快照仓库
func deleteSnapshotRepository(client *elasticsearch.Client, repository string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := esapi.SnapshotDeleteRepositoryRequest{
		Repository: []string{repository},
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("执行删除仓库请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES删除仓库失败: %s", res.Status())
	}

	return nil
}

// getSnapshotRepositories 获取快照仓库列表
func getSnapshotRepositories(client *elasticsearch.Client) ([]RepositoryInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := esapi.SnapshotGetRepositoryRequest{
		Repository: []string{"_all"},
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("执行获取仓库请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES获取仓库失败: %s", res.Status())
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析仓库信息失败: %v", err)
	}

	var repositories []RepositoryInfo
	for name, repo := range response {
		if repoMap, ok := repo.(map[string]interface{}); ok {
			repositories = append(repositories, RepositoryInfo{
				Name:     name,
				Type:     repoMap["type"].(string),
				Settings: repoMap["settings"].(map[string]interface{}),
			})
		}
	}

	return repositories, nil
}

// createSnapshot 创建快照
func createSnapshot(client *elasticsearch.Client, repository, snapshot string, req BackupRequest) error {
	// 构建快照配置
	snapshotConfig := map[string]interface{}{
		"indices":              strings.Join(req.Indices, ","),
		"ignore_unavailable":   req.IgnoreUnavailable,
		"include_global_state": req.IncludeGlobalState,
		"metadata":             req.Metadata,
	}

	// 序列化配置
	configBytes, err := json.Marshal(snapshotConfig)
	if err != nil {
		return fmt.Errorf("序列化快照配置失败: %v", err)
	}

	// 执行创建请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	waitForCompletion := false
	reqES := esapi.SnapshotCreateRequest{
		Repository: repository,
		Snapshot:   snapshot,
		Body:       bytes.NewReader(configBytes),
	}
	reqES.WaitForCompletion = &waitForCompletion // 异步执行

	res, err := reqES.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("执行创建快照请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES创建快照失败: %s", res.Status())
	}

	return nil
}

// deleteSnapshot 删除快照
func deleteSnapshot(client *elasticsearch.Client, repository, snapshot string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := esapi.SnapshotDeleteRequest{
		Repository: repository,
		Snapshot:   []string{snapshot},
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("执行删除快照请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES删除快照失败: %s", res.Status())
	}

	return nil
}

// getSnapshots 获取快照列表
func getSnapshots(client *elasticsearch.Client, repository string) ([]BackupInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := esapi.SnapshotGetRequest{
		Repository: repository,
		Snapshot:   []string{"_all"},
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("执行获取快照请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES获取快照失败: %s", res.Status())
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析快照信息失败: %v", err)
	}

	var snapshots []BackupInfo
	if snapshotsData, ok := response["snapshots"].([]interface{}); ok {
		for _, snap := range snapshotsData {
			if snapMap, ok := snap.(map[string]interface{}); ok {
				snapshot := parseSnapshotInfo(snapMap)
				snapshot.Repository = repository
				snapshots = append(snapshots, snapshot)
			}
		}
	}

	return snapshots, nil
}

// getSnapshotStatus 获取快照状态
func getSnapshotStatus(client *elasticsearch.Client, repository, snapshot string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := esapi.SnapshotStatusRequest{
		Repository: repository,
		Snapshot:   []string{snapshot},
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("执行获取快照状态请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES获取快照状态失败: %s", res.Status())
	}

	// 解析响应
	var status map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("解析快照状态失败: %v", err)
	}

	return status, nil
}

// restoreSnapshot 恢复快照
func restoreSnapshot(client *elasticsearch.Client, repository, snapshot string, req RestoreRequest) error {
	// 构建恢复配置
	restoreConfig := map[string]interface{}{
		"indices":              strings.Join(req.Indices, ","),
		"ignore_unavailable":   req.IgnoreUnavailable,
		"include_global_state": req.IncludeGlobalState,
	}

	// 添加重命名规则
	if req.RenamePattern != "" && req.RenameReplacement != "" {
		restoreConfig["rename_pattern"] = req.RenamePattern
		restoreConfig["rename_replacement"] = req.RenameReplacement
	}

	// 序列化配置
	configBytes, err := json.Marshal(restoreConfig)
	if err != nil {
		return fmt.Errorf("序列化恢复配置失败: %v", err)
	}

	// 执行恢复请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	waitForCompletion := false
	reqES := esapi.SnapshotRestoreRequest{
		Repository: repository,
		Snapshot:   snapshot,
		Body:       bytes.NewReader(configBytes),
	}
	reqES.WaitForCompletion = &waitForCompletion // 异步执行

	res, err := reqES.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("执行恢复快照请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// 读取错误响应体以获取更详细的错误信息
		errorBody, _ := io.ReadAll(res.Body)
		return parseRestoreError(res.Status(), string(errorBody))
	}

	return nil
}

// parseRestoreError 解析恢复快照的错误信息，提供更友好的提示
func parseRestoreError(status, errorBody string) error {
	// 尝试解析JSON错误响应
	var errorResponse map[string]interface{}
	if err := json.Unmarshal([]byte(errorBody), &errorResponse); err == nil {
		if error, ok := errorResponse["error"].(map[string]interface{}); ok {
			if rootCause, ok := error["root_cause"].([]interface{}); ok && len(rootCause) > 0 {
				if firstCause, ok := rootCause[0].(map[string]interface{}); ok {
					if errorType, ok := firstCause["type"].(string); ok {
						if reason, ok := firstCause["reason"].(string); ok {
							return formatUserFriendlyError(errorType, reason, status, errorBody)
						}
					}
				}
			}

			// 如果没有root_cause，尝试直接使用error信息
			if errorType, ok := error["type"].(string); ok {
				if reason, ok := error["reason"].(string); ok {
					return formatUserFriendlyError(errorType, reason, status, errorBody)
				}
			}
		}
	}

	// 如果无法解析JSON，返回原始错误
	return fmt.Errorf("ES恢复快照失败: %s, 错误详情: %s", status, errorBody)
}

// formatUserFriendlyError 格式化用户友好的错误信息
func formatUserFriendlyError(errorType, reason, status, originalError string) error {
	switch errorType {
	case "illegal_state_exception":
		if strings.Contains(reason, "has more than one write index") {
			// 解析别名和写索引信息
			aliasName := extractAliasName(reason)
			writeIndices := extractWriteIndices(reason)

			return fmt.Errorf("恢复失败：别名 [%s] 存在多个写索引冲突。\n\n"+
				"冲突的写索引：%s\n\n"+
				"解决方案：\n"+
				"1. 删除现有冲突的索引（如果数据不再需要）\n"+
				"2. 使用重命名功能恢复索引（在恢复时设置重命名规则）\n"+
				"3. 手动管理别名，确保每个别名只有一个写索引\n\n"+
				"技术详情：%s",
				aliasName, strings.Join(writeIndices, ", "), reason)
		}
		if strings.Contains(reason, "index_already_exists_exception") {
			return fmt.Errorf("恢复失败：目标索引已存在。\n\n"+
				"解决方案：\n"+
				"1. 删除现有索引后重新恢复\n"+
				"2. 使用重命名功能恢复索引（在恢复时设置重命名规则）\n\n"+
				"技术详情：%s", reason)
		}
	case "index_not_found_exception":
		return fmt.Errorf("恢复失败：快照中指定的索引不存在。\n\n"+
			"请检查快照内容，确认要恢复的索引名称是否正确。\n\n"+
			"技术详情：%s", reason)
	case "snapshot_restore_exception":
		if strings.Contains(reason, "cannot restore") {
			return fmt.Errorf("恢复失败：快照数据不完整或损坏。\n\n"+
				"解决方案：\n"+
				"1. 检查快照完整性\n"+
				"2. 尝试使用其他快照进行恢复\n"+
				"3. 重新创建快照后再恢复\n\n"+
				"技术详情：%s", reason)
		}
	case "repository_missing_exception":
		return fmt.Errorf("恢复失败：快照仓库不存在。\n\n"+
			"请确认快照仓库已正确创建且可访问。\n\n"+
			"技术详情：%s", reason)
	case "snapshot_missing_exception":
		return fmt.Errorf("恢复失败：指定的快照不存在。\n\n"+
			"请确认快照名称是否正确，以及快照是否存在于指定仓库中。\n\n"+
			"技术详情：%s", reason)
	}

	// 默认返回原始错误信息，但格式化得更友好
	return fmt.Errorf("恢复快照时发生错误：%s\n\n"+
		"错误类型：%s\n"+
		"状态码：%s\n\n"+
		"如需帮助，请查看技术详情或联系管理员。", reason, errorType, status)
}

// extractAliasName 从错误信息中提取别名名称
func extractAliasName(reason string) string {
	// 示例: "alias [filebeat-7.17.29] has more than one write index"
	start := strings.Index(reason, "alias [")
	if start == -1 {
		return "未知别名"
	}
	start += 7 // len("alias [")
	end := strings.Index(reason[start:], "]")
	if end == -1 {
		return "未知别名"
	}
	return reason[start : start+end]
}

// extractWriteIndices 从错误信息中提取写索引列表
func extractWriteIndices(reason string) []string {
	// 示例: "has more than one write index [restored-filebeat-7.17.29-2025.10.30-000001,filebeat-7.17.29-2025.10.30-000001-restored]"
	start := strings.Index(reason, "write index [")
	if start == -1 {
		return []string{"未知索引"}
	}
	start += 12 // len("write index [")
	end := strings.Index(reason[start:], "]")
	if end == -1 {
		return []string{"未知索引"}
	}
	indicesStr := reason[start : start+end]
	return strings.Split(indicesStr, ",")
}

// getRestoreStatus 获取恢复状态
func getRestoreStatus(client *elasticsearch.Client) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := esapi.SnapshotStatusRequest{
		Repository: "_all",
		Snapshot:   []string{"_all"},
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("执行获取恢复状态请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES获取恢复状态失败: %s", res.Status())
	}

	// 解析响应
	var status map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("解析恢复状态失败: %v", err)
	}

	return status, nil
}

// verifyRepository 验证仓库
func verifyRepository(client *elasticsearch.Client, repositoryName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := esapi.SnapshotVerifyRepositoryRequest{
		Repository: repositoryName,
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("执行验证仓库请求失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// 读取错误响应体以获取更详细的错误信息
		errorBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("ES验证仓库失败: %s, 错误详情: %s", res.Status(), string(errorBody))
	}

	return nil
}

// parseSnapshotInfo 解析快照信息
func parseSnapshotInfo(data map[string]interface{}) BackupInfo {
	info := BackupInfo{}

	if snapshot, ok := data["snapshot"].(string); ok {
		info.Snapshot = snapshot
	}
	if uuid, ok := data["uuid"].(string); ok {
		info.UUID = uuid
	}
	if version, ok := data["version"].(string); ok {
		info.Version = version
	}
	if startTime, ok := data["start_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			info.StartTime = t
		}
	}
	if endTime, ok := data["end_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			info.EndTime = t
		}
	}
	if duration, ok := data["duration_in_millis"].(float64); ok {
		info.Duration = fmt.Sprintf("%.2fs", duration/1000)
	}
	if state, ok := data["state"].(string); ok {
		info.State = state
	}
	if reason, ok := data["reason"].(string); ok {
		info.Reason = reason
	}
	if shards, ok := data["shards"].(map[string]interface{}); ok {
		info.Shards = shards
	}
	if includeGlobalState, ok := data["include_global_state"].(bool); ok {
		info.IncludeGlobalState = includeGlobalState
	}
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		info.Metadata = metadata
	}

	// 提取索引列表
	if indices, ok := data["indices"].([]interface{}); ok {
		for _, idx := range indices {
			if indexName, ok := idx.(string); ok {
				info.Indices = append(info.Indices, indexName)
			}
		}
	}

	return info
}
