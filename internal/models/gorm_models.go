package models

import (
	"time"
)

// InstanceType 实例类型模型
type InstanceType struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id;type:int unsigned" json:"id"` // 明确指定为int unsigned类型
	TypeName    string    `gorm:"unique;not null;column:type_name;size:100" json:"type_name"`
	Description string    `gorm:"column:description" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (InstanceType) TableName() string {
	return "instance_types"
}

// Instance 实例模型
type Instance struct {
	ID             uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`                                     // 使用uint匹配UNSIGNED INT
	InstanceTypeID uint      `gorm:"not null;index;column:instance_type_id;type:int unsigned" json:"instance_type_id"` // 明确指定为int unsigned类型
	Name           string    `gorm:"not null;column:name;size:255" json:"name"`
	Address        string    `gorm:"not null;column:address;size:500" json:"address"`
	HttpsEnabled   bool      `gorm:"default:false;index;column:https_enabled" json:"https_enabled"`
	SkipSslVerify  bool      `gorm:"default:false;column:skip_ssl_verify" json:"skip_ssl_verify"`
	Status         string    `gorm:"default:'active';index;column:status" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// AuthConfig 统一认证配置模型
type AuthConfig struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`                // 使用uint匹配UNSIGNED INT
	ResourceType string    `gorm:"not null;index;column:resource_type" json:"resource_type"`    // instance或cluster
	ResourceID   uint      `gorm:"not null;index;column:resource_id" json:"resource_id"`        // 使用uint匹配UNSIGNED INT
	ResourceName string    `gorm:"not null;column:resource_name;size:255" json:"resource_name"` // 资源名称
	AuthType     string    `gorm:"not null;index;column:auth_type" json:"auth_type"`
	ConfigKey    string    `gorm:"not null;column:config_key;size:100" json:"config_key"`
	ConfigValue  string    `gorm:"column:config_value" json:"config_value"`
	IsEncrypted  bool      `gorm:"default:true;column:is_encrypted" json:"is_encrypted"`
	Status       string    `gorm:"default:'active';column:status" json:"status"` // 配置状态
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (AuthConfig) TableName() string {
	return "auth_configs"
}

// ConnectionTest 连接测试记录模型
type ConnectionTest struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`             // 使用uint匹配UNSIGNED INT
	ResourceType string    `gorm:"not null;index;column:resource_type" json:"resource_type"` // instance或cluster
	ResourceID   uint      `gorm:"not null;index;column:resource_id" json:"resource_id"`     // 使用uint匹配UNSIGNED INT
	TestResult   string    `gorm:"column:test_result" json:"test_result"`
	ResponseTime *int      `gorm:"column:response_time" json:"response_time"` // 使用指针允许NULL值
	ErrorMessage string    `gorm:"column:error_message" json:"error_message"`
	TestedAt     time.Time `gorm:"index;column:tested_at" json:"tested_at"`
}

// TableName 指定表名
func (ConnectionTest) TableName() string {
	return "connection_tests"
}
