package dal

import (
	"encoding/json"
	"time"
)

// ResourceDetail 统一资源详情视图模型
type ResourceDetail struct {
	ResourceType    string    `json:"resource_type"`                                                     // 资源类型：instance或cluster
	ResourceID      uint      `json:"resource_id"`                                                       // 资源ID（使用uint匹配UNSIGNED INT）
	ResourceName    string    `json:"resource_name"`                                                     // 资源名称
	Status          string    `json:"status"`                                                            // 状态
	CreatedAt       time.Time `json:"created_at"`                                                        // 创建时间
	UpdatedAt       time.Time `json:"updated_at"`                                                        // 更新时间
	AuthConfigs     string    `gorm:"column:auth_configs" json:"auth_configs"`                           // 认证配置JSON数组
	AuthTypeDesc    string    `json:"auth_type_desc"`                                                    // 认证类型描述
	Address         *string   `gorm:"column:connection_endpoint" json:"address,omitempty"`               // 地址（仅实例）
	HttpsEnabled    *bool     `gorm:"column:secure_connection" json:"https_enabled,omitempty"`           // HTTPS启用（仅实例）
	SkipSslVerify   *bool     `gorm:"column:ssl_verification_disabled" json:"skip_ssl_verify,omitempty"` // 跳过SSL验证（仅实例）
	TypeName        string    `gorm:"column:resource_subtype" json:"type_name"`                          // 类型名称
	TypeDescription string    `gorm:"column:subtype_description" json:"type_description"`                // 类型描述
}

// TableName 指定表名
func (ResourceDetail) TableName() string {
	return "resource_details"
}

// GetAuthConfigs 获取认证配置数组
func (rd *ResourceDetail) GetAuthConfigs() []map[string]interface{} {
	if rd.AuthConfigs == "" {
		return nil
	}

	var configs []map[string]interface{}
	if err := json.Unmarshal([]byte(rd.AuthConfigs), &configs); err != nil {
		return nil
	}
	return configs
}

// GetAuthConfigByKey 根据键获取认证配置值
func (rd *ResourceDetail) GetAuthConfigByKey(key string) (map[string]interface{}, bool) {
	configs := rd.GetAuthConfigs()
	if configs == nil {
		return nil, false
	}

	for _, config := range configs {
		if configKey, ok := config["config_key"].(string); ok && configKey == key {
			return config, true
		}
	}
	return nil, false
}

// GetAuthConfigValue 根据键获取认证配置值
func (rd *ResourceDetail) GetAuthConfigValue(key string) (string, bool) {
	config, exists := rd.GetAuthConfigByKey(key)
	if !exists {
		return "", false
	}

	if value, ok := config["config_value"].(string); ok {
		return value, true
	}
	return "", false
}

// ResourceFilter 资源过滤器
type ResourceFilter struct {
	ResourceType *string `json:"resource_type"` // 资源类型过滤
	Status       *string `json:"status"`        // 状态过滤
	TypeName     *string `json:"type_name"`     // 类型名称过滤
}

// ResourceList 资源列表
type ResourceList struct {
	Resources []ResourceDetail `json:"resources"`
	Total     int              `json:"total"`
}
