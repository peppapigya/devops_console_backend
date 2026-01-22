package common

import (
	"crypto/tls"
	"devops-console-backend/models"
	"devops-console-backend/models/request"
	"devops-console-backend/repositories"
	"devops-console-backend/utils"
	"devops-console-backend/utils/logs"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 测试结果常量
const (
	TestResultSuccess = "success"
	TestResultFailure = "failure"
	TestResultTimeout = "timeout"
)

// TestConnection 测试连接接口
// @Summary 测试实例连接
// @Description 测试指定实例的连接状态和响应时间
// @Tags instance
// @Accept json
// @Produce json
// @Param request body request.ConnectionTestRequest true "连接测试请求"
// @Success 200 {object} common.ReturnData "成功"
// @Failure 400 {object} common.ReturnData "请求参数错误"
// @Failure 500 {object} common.ReturnData "服务器内部错误"
// @Router /api/instance/test-connection [post]
func TestConnection(r *gin.Context) {
	helper := utils.NewResponseHelper(r)
	var req request.ConnectionTestRequest

	// 绑定请求参数
	if err := r.ShouldBindJSON(&req); err != nil {
		helper.LogAndBadRequest("请求参数绑定失败", map[string]interface{}{"error": err.Error()})
		return
	}

	// 验证实例是否存在
	instance, err := validateInstanceExists(req.InstanceID)
	if err != nil {
		helper.LogAndBadRequest("实例验证失败", map[string]interface{}{
			"instance_id": req.InstanceID,
			"error":       err.Error(),
		})
		return
	}

	// 获取认证配置
	authConfig, err := getAuthConfig(req.InstanceID)
	if err != nil {
		logs.Warning(map[string]interface{}{
			"instance_id": req.InstanceID,
			"error":       err.Error(),
		}, "获取认证配置失败，使用无认证方式")
		authConfig = nil
	}

	// 执行连接测试
	startTime := time.Now()
	result, responseTime, errMsg := performConnectionTest(instance, authConfig)
	testDuration := time.Since(startTime)

	// 保存测试记录
	var responseTimePtr *int
	if responseTime > 0 {
		responseTimeInt := int(responseTime)
		responseTimePtr = &responseTimeInt
	}

	testRecord := models.ConnectionTest{
		ResourceType: "instance",
		ResourceID:   uint(req.InstanceID),
		TestResult:   string(result),
		ResponseTime: responseTimePtr,
		ErrorMessage: errMsg,
		TestedAt:     time.Now(),
	}

	if err := saveTestRecord(&testRecord); err != nil {
		logs.Error(map[string]interface{}{
			"instance_id": req.InstanceID,
			"error":       err.Error(),
		}, "保存测试记录失败")
	}

	// 返回测试结果
	responseData := map[string]interface{}{
		"instance_id":   req.InstanceID,
		"instance_name": instance.Name,
		"test_result":   result,
		"response_time": responseTime,
		"error_message": errMsg,
		"tested_at":     testRecord.TestedAt,
		"test_duration": testDuration.Milliseconds(),
	}

	logs.Info(map[string]interface{}{
		"instance_id":   req.InstanceID,
		"instance_name": instance.Name,
		"test_result":   result,
		"response_time": responseTime,
	}, "连接测试完成")

	if result == TestResultSuccess {
		helper.SuccessWithData("连接测试成功", "test_result", responseData)
	} else {
		errorMsg := "连接测试失败"
		if errMsg != "" {
			errorMsg += ": " + errMsg
		}
		helper.InternalError(errorMsg)
	}
}

// validateInstanceExists 验证实例是否存在
func validateInstanceExists(instanceID uint) (*models.Instance, error) {
	instanceRepo := repositories.NewInstanceRepository()
	instance, err := instanceRepo.GetByID(instanceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("实例不存在")
		}
		return nil, fmt.Errorf("查询实例失败: %w", err)
	}
	return instance, nil
}

// getAuthConfig 获取认证配置
func getAuthConfig(instanceID uint) (*models.AuthConfig, error) {
	authConfigRepo := repositories.NewAuthConfigRepository()
	authConfigs, err := authConfigRepo.GetByInstanceID(instanceID)
	if err != nil {
		return nil, fmt.Errorf("查询认证配置失败: %w", err)
	}
	if len(authConfigs) == 0 {
		return nil, nil // 无认证配置
	}
	return &authConfigs[0], nil
}

// performConnectionTest 执行连接测试
func performConnectionTest(instance *models.Instance, authConfig *models.AuthConfig) (string, int64, string) {
	// 检查是否为 Kubernetes 类型
	instanceDetailRepo := repositories.NewInstanceDetailRepository()
	instanceDetail, err := instanceDetailRepo.GetByID(instance.ID)

	// 添加调试日志
	logs.Info(map[string]interface{}{
		"instance_id":   instance.ID,
		"instance_name": instance.Name,
		"type_name": func() string {
			if err == nil {
				return instanceDetail.TypeName
			}
			return "unknown"
		}(),
	}, "开始连接测试")

	if err == nil && (instanceDetail.TypeName == "Kubernetes" || instanceDetail.TypeName == "kubernetes") {
		// 对于 Kubernetes，使用特殊的测试方法
		logs.Info(map[string]interface{}{
			"instance_id":   instance.ID,
			"instance_name": instance.Name,
		}, "使用 Kubernetes 连接测试方法")
		return performKubernetesConnectionTest(instance, authConfig)
	}

	// 构建请求URL
	protocol := "http"
	if instance.HttpsEnabled {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s", protocol, instance.Address)

	// 创建HTTP客户端，配置SSL跳过验证
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 如果是HTTPS且需要跳过SSL验证
	if instance.HttpsEnabled && instance.SkipSslVerify {
		// 创建自定义Transport来跳过SSL验证
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return TestResultFailure, 0, fmt.Sprintf("创建请求失败: %s", err.Error())
	}

	// 设置认证信息
	if authConfig != nil && authConfig.AuthType != "none" {
		switch authConfig.AuthType {
		case "basic":
			if authConfig.ConfigValue != "" {
				// 解析基础认证配置JSON
				var basicAuth struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}
				if err := json.Unmarshal([]byte(authConfig.ConfigValue), &basicAuth); err == nil {
					if basicAuth.Username != "" && basicAuth.Password != "" {
						req.SetBasicAuth(basicAuth.Username, basicAuth.Password)
					}
				} else {
					logs.Warning(map[string]interface{}{
						"instance_id":  instance.ID,
						"config_value": authConfig.ConfigValue,
						"error":        err.Error(),
					}, "解析基础认证配置失败")
				}
			}
		case "api_key":
			if authConfig.ConfigValue != "" {
				// 解析API密钥配置JSON
				var apiKeyAuth struct {
					ApiKey string `json:"apiKey"`
					ApiId  string `json:"apiId"`
				}
				if err := json.Unmarshal([]byte(authConfig.ConfigValue), &apiKeyAuth); err == nil {
					if apiKeyAuth.ApiKey != "" {
						// 使用API Key作为认证
						req.Header.Set("Authorization", "ApiKey "+apiKeyAuth.ApiKey)
					}
				} else {
					// 如果解析失败，直接使用原始值作为API Key
					req.Header.Set("Authorization", "ApiKey "+authConfig.ConfigValue)
				}
			}
		case "token":
			if authConfig.ConfigValue != "" {
				// 解析令牌配置JSON
				var tokenAuth struct {
					Token string `json:"token"`
				}
				if err := json.Unmarshal([]byte(authConfig.ConfigValue), &tokenAuth); err == nil {
					if tokenAuth.Token != "" {
						req.Header.Set("Authorization", "Bearer "+tokenAuth.Token)
					}
				} else {
					// 如果解析失败，直接使用原始值作为Token
					req.Header.Set("Authorization", "Bearer "+authConfig.ConfigValue)
				}
			}
		default:
			logs.Warning(map[string]interface{}{
				"instance_id": instance.ID,
				"auth_type":   authConfig.AuthType,
			}, "不支持的认证类型，将使用无认证方式")
		}
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "devops-console/1.0")

	// 执行请求
	startTime := time.Now()
	resp, err := client.Do(req)
	responseTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if err.Error() == "context deadline exceeded" {
			return TestResultTimeout, responseTime, "连接超时"
		}
		return TestResultFailure, responseTime, fmt.Sprintf("连接失败: %s", err.Error())
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return TestResultSuccess, responseTime, ""
	}

	return TestResultFailure, responseTime, fmt.Sprintf("服务器返回错误状态码: %d", resp.StatusCode)
}

// performKubernetesConnectionTest 执行 Kubernetes 连接测试
func performKubernetesConnectionTest(instance *models.Instance, authConfig *models.AuthConfig) (string, int64, string) {
	startTime := time.Now()

	// 检查认证配置
	if authConfig == nil {
		return TestResultFailure, 0, "缺少 Kubernetes 认证配置"
	}

	if authConfig.AuthType != "kubeconfig" {
		logs.Warning(map[string]interface{}{
			"instance_id": instance.ID,
			"auth_type":   authConfig.AuthType,
		}, "不支持的认证类型，将使用无认证方式")
		return TestResultFailure, 0, "不支持的认证类型: " + authConfig.AuthType
	}

	if authConfig.ConfigValue == "" {
		return TestResultFailure, 0, "Kubernetes kubeconfig 配置为空"
	}

	// 解析 kubeconfig 配置
	var kubeconfig struct {
		Config string `json:"config"`
	}
	if err := json.Unmarshal([]byte(authConfig.ConfigValue), &kubeconfig); err != nil {
		return TestResultFailure, 0, "解析 kubeconfig 配置失败: " + err.Error()
	}

	if kubeconfig.Config == "" {
		// 如果没有嵌套的config字段，直接使用configValue作为kubeconfig内容
		kubeconfig.Config = authConfig.ConfigValue
	}

	// TODO: 实际的 Kubernetes 连接测试
	// 这里应该使用 client-go 库来验证 kubeconfig 并测试连接
	// 暂时简化处理，验证配置不为空即认为成功
	responseTime := time.Since(startTime).Milliseconds()

	// 简单验证 kubeconfig 格式
	if len(kubeconfig.Config) > 50 && (strings.Contains(kubeconfig.Config, "apiVersion") ||
		strings.Contains(kubeconfig.Config, "clusters") ||
		strings.Contains(kubeconfig.Config, "current-context")) {
		return TestResultSuccess, responseTime, "Kubernetes 配置验证通过"
	}

	return TestResultFailure, responseTime, "Kubernetes 配置格式不正确"
}

// saveTestRecord 保存测试记录
func saveTestRecord(testRecord *models.ConnectionTest) error {
	connectionTestRepo := repositories.NewConnectionTestRepository()
	if err := connectionTestRepo.Create(testRecord); err != nil {
		return fmt.Errorf("保存测试记录失败: %w", err)
	}

	return nil
}

// GetTestHistory 获取测试历史记录
// @Summary 获取测试历史记录
// @Description 获取指定实例的连接测试历史记录
// @Tags instance
// @Accept json
// @Produce json
// @Param instance_id query int true "实例ID"
// @Param page_size query int false "每页记录数" default(10)
// @Success 200 {object} common.ReturnData "成功"
// @Failure 400 {object} common.ReturnData "请求参数错误"
// @Failure 500 {object} common.ReturnData "服务器内部错误"
// @Router /api/instance/test-history [get]
func GetTestHistory(r *gin.Context) {
	helper := utils.NewResponseHelper(r)

	// 获取查询参数
	instanceIDStr := r.Query("instance_id")
	pageSizeStr := r.Query("page_size")

	// 参数验证和转换
	if instanceIDStr == "" {
		helper.BadRequest("instance_id 参数不能为空")
		return
	}

	instanceID64, err := strconv.ParseInt(instanceIDStr, 10, 64)
	instanceID := uint(instanceID64)
	if err != nil {
		helper.BadRequest("instance_id 参数格式错误")
		return
	}

	pageSize := 10
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// 使用GORM查询测试记录
	connectionTestRepo := repositories.NewConnectionTestRepository()
	tests, err := connectionTestRepo.GetByInstanceID(instanceID, pageSize)
	if err != nil {
		helper.DatabaseError("查询测试记录失败: " + err.Error())
		return
	}

	logs.Debug(nil, "查询测试历史记录成功")
	helper.Success("查询成功", map[string]interface{}{
		"tests":     tests,
		"page_size": pageSize,
		"test_list": tests,
	})
}

// GetTodayTestStats 获取今日测试统计
// @Summary 获取今日测试统计
// @Description 获取今日所有实例的连接测试统计信息
// @Tags instance
// @Accept json
// @Produce json
// @Success 200 {object} common.ReturnData "成功"
// @Failure 500 {object} common.ReturnData "服务器内部错误"
// @Router /api/instance/today-test-stats [get]
func GetTodayTestStats(r *gin.Context) {
	helper := utils.NewResponseHelper(r)

	// 使用GORM查询今日统计信息
	connectionTestRepo := repositories.NewConnectionTestRepository()
	stats, err := connectionTestRepo.GetTodayStats()
	if err != nil {
		helper.DatabaseError("查询今日测试统计失败: " + err.Error())
		return
	}

	totalTests := stats["total_tests"].(int64)
	instanceTests := stats["instance_tests"].([]struct {
		InstanceID uint   `json:"instance_id"`
		Name       string `json:"name"`
		Count      int64  `json:"count"`
	})

	instanceTestCounts := make(map[string]int64)
	for _, instanceTest := range instanceTests {
		instanceTestCounts[strconv.FormatInt(int64(instanceTest.InstanceID), 10)] = instanceTest.Count
	}

	logs.Debug(map[string]interface{}{
		"total_tests":          totalTests,
		"instance_test_counts": instanceTestCounts,
	}, "查询今日测试统计成功")

	helper.Success("查询成功", map[string]interface{}{
		"total_tests":          totalTests,
		"instance_test_counts": instanceTestCounts,
		"date":                 time.Now().Format("2006-01-02"),
	})
}
