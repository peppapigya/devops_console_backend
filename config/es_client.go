package config

import (
	"context"
	"crypto/tls"
	"devops-console-backend/database"
	"devops-console-backend/models"
	"devops-console-backend/pkg/utils/logs"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
)

var esClientsMutex sync.RWMutex

var EsClients = make(map[uint]*elasticsearch.Client)

// InitEsClients 初始化 es 客户端
func InitEsClients() {
	// 使用GORM查询实例详情视图
	instanceDetails, err := getElasticsearchInstances()
	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "获取Elasticsearch实例列表失败")
		return
	}

	// 使用认证配置创建 es 客户端
	for _, instanceDetail := range instanceDetails {
		LockEsClients() // 加锁
		client, err := createEsClient(instanceDetail)
		if err != nil {
			logs.Error(map[string]interface{}{"instance_id": instanceDetail.ResourceID, "error": err.Error()}, "创建Elasticsearch客户端失败")
			UnlockEsClients() // 出错时也需要解锁
			continue
		}

		EsClients[instanceDetail.ResourceID] = client // 将客户端存储到 esClients
		logs.Info(map[string]interface{}{"instance_id": instanceDetail.ResourceID}, "Elasticsearch客户端创建成功")

		UnlockEsClients() // 解锁
	}
}

// 创建 es 客户端
func createEsClient(instanceDetail models.ResourceDetail) (*elasticsearch.Client, error) {
	// 从Address字段获取地址，需要先转换为字符串
	var addr string
	if instanceDetail.Address != nil {
		addr = *instanceDetail.Address
	} else {
		return nil, fmt.Errorf("实例地址为空")
	}

	if instanceDetail.HttpsEnabled != nil && *instanceDetail.HttpsEnabled == true {
		addr = "https://" + addr
	} else {
		addr = "http://" + addr
	}

	// 配置客户端的参数
	cfg := elasticsearch.Config{
		Addresses: []string{addr}, // 集群的地址
	}

	// 解析认证配置
	authConfigs := parseAuthConfigs(string(instanceDetail.AuthConfigs))

	for authType, configValue := range authConfigs {
		switch authType {
		case "basic":
			if configValue.ConfigValue != "" {
				raw := strings.TrimSpace(configValue.ConfigValue)
				if strings.HasPrefix(raw, "{") {
					var basicMap map[string]string
					if err := json.Unmarshal([]byte(raw), &basicMap); err == nil {
						cfg.Username = basicMap["username"]
						cfg.Password = basicMap["password"]
						break
					}
				}

				cfg.Username = configValue.ConfigKey
				cfg.Password = configValue.ConfigValue
			} else {
				cfg.Username = configValue.ConfigKey
			}

		case "api_key":
			if configValue.ConfigValue != "" {
				cfg.APIKey = configValue.ConfigValue // API 密钥
			}
		}
	}

	// 跳过证书验证并设置超时
	skipSSL := false
	if instanceDetail.SkipSslVerify != nil {
		skipSSL = *instanceDetail.SkipSslVerify
	}
	cfg.Transport = &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: skipSSL},
		ResponseHeaderTimeout: 10 * time.Second,
	}

	// 创建 Elasticsearch 客户端
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %v", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("failed to get response from Elasticsearch: %v", res.Status())
	}

	return client, nil
}

// 解析认证配置
func parseAuthConfigs(authConfigsStr string) map[string]models.AuthConfig {
	authConfigs := make(map[string]models.AuthConfig)

	// 认证配置是以JSON格式存储的
	if authConfigsStr == "" {
		return authConfigs
	}

	// 尝试解析为JSON格式
	var jsonConfigs map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(authConfigsStr), &jsonConfigs); err == nil {
		// JSON格式解析成功
		for authType, configData := range jsonConfigs {
			if configValue, ok := configData["config_value"].(string); ok {
				authConfigs[authType] = models.AuthConfig{
					ConfigKey:   authType,
					ConfigValue: configValue,
				}
			}
		}
		return authConfigs
	}

	// 如果JSON解析失败，尝试按逗号分隔的键值对格式解析
	configPairs := splitTopLevelCommas(authConfigsStr)
	for _, pair := range configPairs {
		colonIndex := strings.Index(pair, ":")
		if colonIndex <= 0 {
			continue
		}
		key := pair[:colonIndex]
		value := pair[colonIndex+1:]

		if strings.HasPrefix(strings.TrimSpace(value), "{") {
			authConfigs[key] = models.AuthConfig{
				ConfigKey:   key,
				ConfigValue: value,
			}
			continue
		}

		// 普通 key:value 格式
		authConfigs[key] = models.AuthConfig{
			ConfigKey:   key,
			ConfigValue: value,
		}
	}

	return authConfigs
}

func splitTopLevelCommas(s string) []string {
	var parts []string
	last := 0
	depth := 0
	inQuotes := false
	var prevRune rune
	for i, r := range s {
		if r == '"' && prevRune != '\\' {
			inQuotes = !inQuotes
		}
		if !inQuotes {
			if r == '{' || r == '[' || r == '(' {
				depth++
			} else if r == '}' || r == ']' || r == ')' {
				if depth > 0 {
					depth--
				}
			} else if r == ',' && depth == 0 {
				parts = append(parts, strings.TrimSpace(s[last:i]))
				last = i + 1
			}
		}
		prevRune = r
	}
	if last <= len(s)-1 {
		parts = append(parts, strings.TrimSpace(s[last:]))
	}
	return parts
}

// 关闭 es 客户端连接
func closeEsClient(client *elasticsearch.Client) {
	if client.Transport != nil {
		// 关闭连接
		if esTransport, ok := client.Transport.(*elastictransport.Client); ok {
			// 使用反射
			rv := reflect.ValueOf(esTransport).Elem()
			transportField := rv.FieldByName("transport")
			if transportField.IsValid() && !transportField.IsZero() {
				// 获取http.Transport
				transportValue := transportField.Interface()
				if httpTransport, ok := transportValue.(*http.Transport); ok {
					//关闭空闲连接
					httpTransport.CloseIdleConnections()
					log.Println("Elasticsearch 客户端连接已关闭")
				} else {
					log.Println("无法访问底层的 http.Transport")
				}
			} else {
				log.Println("无法访问 transport 字段")
			}
		} else {
			log.Println("无法访问正确的 Transport 类型")
		}
	}
}

// CreateEsClient 封装内部
func CreateEsClient(authConfig models.AuthConfig) (*elasticsearch.Client, error) {

	if authConfig.ResourceID == 0 {
		return nil, fmt.Errorf("invalid resource id: %d", authConfig.ResourceID)
	}

	instanceDetail, err := getInstanceDetailByID(authConfig.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query instance detail: %v", err)
	}

	// 使用内部 createEsClient 创建客户端
	client, err := createEsClient(*instanceDetail)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getEsClient(instanceID uint) (*elasticsearch.Client, bool) {
	esClientsMutex.RLock()
	defer esClientsMutex.RUnlock()
	c, ok := EsClients[instanceID]
	return c, ok
}

// GetEsClient 封装内部，添加按需初始化逻辑
func GetEsClient(instanceID uint) (*elasticsearch.Client, bool) {
	// 先尝试从缓存中获取
	client, exists := getEsClient(instanceID)
	if exists {
		return client, true
	}

	// 如果不存在，则按需初始化
	logs.Info(map[string]interface{}{"instance_id": instanceID}, "Elasticsearch客户端不存在，开始按需初始化")

	// 获取实例详情
	instanceDetail, err := getInstanceDetailByID(instanceID)
	if err != nil {
		logs.Error(map[string]interface{}{"instance_id": instanceID, "error": err.Error()}, "获取实例详情失败")
		return nil, false
	}

	// 创建客户端
	client, err = createEsClient(*instanceDetail)
	if err != nil {
		logs.Error(map[string]interface{}{"instance_id": instanceID, "error": err.Error()}, "创建Elasticsearch客户端失败")
		return nil, false
	}

	// 存储到缓存
	LockEsClients()
	EsClients[instanceID] = client
	UnlockEsClients()

	logs.Info(map[string]interface{}{"instance_id": instanceID}, "Elasticsearch客户端按需初始化成功")
	return client, true
}

// CloseEsClient 封装内部
func CloseEsClient(client *elasticsearch.Client) {
	closeEsClient(client)
}

// LockEsClients 加锁
func LockEsClients() {
	esClientsMutex.Lock()
}

// UnlockEsClients 解锁
func UnlockEsClients() {
	esClientsMutex.Unlock()
}

// SafeSetEsClient 在加锁环境下设置客户端（需外部先调用 LockEsClients）
func SafeSetEsClient(instanceID uint, client *elasticsearch.Client) {
	EsClients[instanceID] = client
	logs.Info(map[string]interface{}{"instance_id": instanceID}, "已创建/更新 Elasticsearch 客户端")
}

// SafeDeleteEsClient 在加锁环境下删除客户端（需外部先调用 LockEsClients）
func SafeDeleteEsClient(instanceID uint) {
	delete(EsClients, instanceID)
}

// getElasticsearchInstances 获取所有Elasticsearch实例
func getElasticsearchInstances() ([]models.ResourceDetail, error) {
	// 使用原生SQL查询避免循环依赖
	var instanceDetails []models.ResourceDetail
	err := database.GORMDB.Where("resource_type = ? AND resource_subtype = ?", "instance", "elasticsearch").Find(&instanceDetails).Error
	return instanceDetails, err
}

// getInstanceDetailByID 根据ID获取实例详情
func getInstanceDetailByID(id uint) (*models.ResourceDetail, error) {
	var instanceDetail models.ResourceDetail
	err := database.GORMDB.Where("resource_id = ? AND resource_type = ?", id, "instance").First(&instanceDetail).Error
	if err != nil {
		return nil, err
	}
	return &instanceDetail, nil
}
