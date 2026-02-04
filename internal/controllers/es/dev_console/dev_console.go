package dev_console

import (
	"bytes"
	"context"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/logs"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
)

// DevConsoleRequest 开发者控制台请求结构体
type DevConsoleRequest struct {
	InstanceID uint              `json:"instance_id" binding:"required"`
	Method     string            `json:"method" binding:"required"`
	Path       string            `json:"path" binding:"required"`
	Body       string            `json:"body"`
	Params     map[string]string `json:"params"`
}

// DevConsoleResponse 开发者控制台响应结构体
type DevConsoleResponse struct {
	Status   int               `json:"status"`
	Headers  map[string]string `json:"headers"`
	Body     interface{}       `json:"body"`
	RawBody  string            `json:"raw_body"`
	Duration int64             `json:"duration_ms"`
}

// DevConsoleHandler 处理开发者模式的通用请求（通过查询参数方式）
func DevConsoleHandler(c *gin.Context) {
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
	client, exists := configs.GetEsClient(instanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 获取请求参数
	method := c.Request.Method
	path := c.Param("path")

	// 读取请求体
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		// 重新设置Body以便后续处理
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 获取查询参数（过滤掉instance_id，因为这是内部使用的参数）
	params := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 && key != "instance_id" {
			params[key] = values[0]
		}
	}

	// 执行ES请求
	response, err := executeDevConsoleRequest(client, method, path, string(bodyBytes), params)
	if err != nil {
		helper.DatabaseError("执行ES请求失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{
		"instance_id": instanceID,
		"method":      method,
		"path":        path,
		"status":      response.Status,
		"duration":    response.Duration,
	}, "开发者控制台请求成功")

	// 不设置原始响应头，避免Content-Length冲突
	// 只返回标准格式的响应
	if response.Status >= 200 && response.Status < 300 {
		helper.SuccessWithData("请求成功", "data", response.Body)
	} else {
		helper.Error(response.Status, fmt.Sprintf("请求失败: %s", response.RawBody))
	}
}

// ExecuteDevConsoleRequest 执行开发者控制台请求（通过JSON请求体方式）
func ExecuteDevConsoleRequest(c *gin.Context) {
	helper := utils.NewResponseHelper(c)

	var req DevConsoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest("请求参数绑定失败: " + err.Error())
		return
	}

	// 获取ES客户端
	client, exists := configs.GetEsClient(req.InstanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 执行ES请求
	response, err := executeDevConsoleRequest(client, req.Method, req.Path, req.Body, req.Params)
	if err != nil {
		helper.DatabaseError("执行ES请求失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{
		"instance_id": req.InstanceID,
		"method":      req.Method,
		"path":        req.Path,
		"status":      response.Status,
		"duration":    response.Duration,
	}, "开发者控制台请求成功")

	helper.SuccessWithData("执行ES请求成功", "response", response)
}

// executeDevConsoleRequest 执行ES请求的核心逻辑
func executeDevConsoleRequest(client *elasticsearch.Client, method, path, body string, params map[string]string) (*DevConsoleResponse, error) {
	startTime := time.Now()

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 构建请求体
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	// 构建请求参数
	var headerParams map[string]string
	if params != nil && len(params) > 0 {
		headerParams = params
	}

	// 使用ES客户端的底层Transport执行请求
	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, method, path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置查询参数
	if headerParams != nil {
		q := httpReq.URL.Query()
		for k, v := range headerParams {
			q.Set(k, v)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")

	// 使用client.Transport执行请求
	httpRes, err := client.Transport.Perform(httpReq)
	if err != nil {
		return nil, fmt.Errorf("执行HTTP请求失败: %v", err)
	}
	defer httpRes.Body.Close()

	// 转换为ES API响应格式
	res := &esapi.Response{
		StatusCode: httpRes.StatusCode,
		Body:       httpRes.Body,
		Header:     httpRes.Header,
	}

	// 读取响应体
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("读取ES响应失败: %v", err)
	}

	// 构建响应头
	headers := make(map[string]string)
	for key, values := range res.Header {
		if len(values) > 0 {
			// 过滤掉一些可能导致问题的响应头
			if key != "Transfer-Encoding" && key != "Connection" {
				headers[key] = values[0]
			}
		}
	}

	// 尝试解析JSON响应体
	var parsedBody interface{}
	contentType := headers["Content-Type"]
	if contentType == "" {
		contentType = "application/json" // 默认为JSON
	}

	if strings.Contains(contentType, "application/json") {
		if err := json.Unmarshal(bodyBytes, &parsedBody); err != nil {
			// 如果解析失败，使用原始字符串
			parsedBody = string(bodyBytes)
		}
	} else {
		parsedBody = string(bodyBytes)
	}

	// 构建响应
	response := &DevConsoleResponse{
		Status:   res.StatusCode,
		Headers:  headers,
		Body:     parsedBody,
		RawBody:  string(bodyBytes),
		Duration: time.Since(startTime).Milliseconds(),
	}

	return response, nil
}

// ValidateInstance 验证实例是否可用
func ValidateInstance(c *gin.Context) {
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
	client, exists := configs.GetEsClient(instanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 测试连接
	err = testConnection(client)
	if err != nil {
		helper.DatabaseError("实例连接测试失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{
		"instance_id": instanceID,
	}, "实例连接测试成功")
	helper.Success("实例连接正常")
}

// testConnection 测试ES连接
func testConnection(client *elasticsearch.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := esapi.InfoRequest{}
	res, err := req.Do(ctx, client)
	if err != nil {
		return fmt.Errorf("执行连接测试失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES返回错误: %s", res.Status())
	}

	return nil
}

// GetInstanceInfo 获取实例详细信息
func GetInstanceInfo(c *gin.Context) {
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
	client, exists := configs.GetEsClient(instanceID)
	if !exists {
		helper.BadRequest("实例不存在或未初始化")
		return
	}

	// 获取实例信息
	info, err := getInstanceInfo(client)
	if err != nil {
		helper.DatabaseError("获取实例信息失败: " + err.Error())
		return
	}

	logs.Info(map[string]interface{}{
		"instance_id": instanceID,
	}, "获取实例信息成功")
	helper.SuccessWithData("获取实例信息成功", "instance_info", info)
}

// getInstanceInfo 获取实例详细信息
func getInstanceInfo(client *elasticsearch.Client) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 并发获取多种信息
	infoCh := make(chan map[string]interface{}, 1)
	healthCh := make(chan map[string]interface{}, 1)
	statsCh := make(chan map[string]interface{}, 1)
	errCh := make(chan error, 3)

	// 获取基本信息
	go func() {
		req := esapi.InfoRequest{}
		res, err := req.Do(ctx, client)
		if err != nil {
			errCh <- fmt.Errorf("获取基本信息失败: %v", err)
			return
		}
		defer res.Body.Close()

		if res.IsError() {
			errCh <- fmt.Errorf("ES返回错误: %s", res.Status())
			return
		}

		var info map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
			errCh <- fmt.Errorf("解析基本信息失败: %v", err)
			return
		}
		infoCh <- info
	}()

	// 获取健康状态
	go func() {
		req := esapi.ClusterHealthRequest{}
		res, err := req.Do(ctx, client)
		if err != nil {
			errCh <- fmt.Errorf("获取健康状态失败: %v", err)
			return
		}
		defer res.Body.Close()

		if res.IsError() {
			errCh <- fmt.Errorf("ES返回错误: %s", res.Status())
			return
		}

		var health map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
			errCh <- fmt.Errorf("解析健康状态失败: %v", err)
			return
		}
		healthCh <- health
	}()

	// 获取统计信息
	go func() {
		req := esapi.ClusterStatsRequest{}
		res, err := req.Do(ctx, client)
		if err != nil {
			errCh <- fmt.Errorf("获取统计信息失败: %v", err)
			return
		}
		defer res.Body.Close()

		if res.IsError() {
			errCh <- fmt.Errorf("ES返回错误: %s", res.Status())
			return
		}

		var stats map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
			errCh <- fmt.Errorf("解析统计信息失败: %v", err)
			return
		}
		statsCh <- stats
	}()

	// 收集结果
	var info, health, stats map[string]interface{}
	for i := 0; i < 3; i++ {
		select {
		case data := <-infoCh:
			info = data
		case data := <-healthCh:
			health = data
		case data := <-statsCh:
			stats = data
		case err := <-errCh:
			return nil, err
		case <-ctx.Done():
			return nil, fmt.Errorf("获取实例信息超时")
		}
	}

	// 组合返回结果
	result := map[string]interface{}{
		"info":   info,
		"health": health,
		"stats":  stats,
	}

	return result, nil
}
