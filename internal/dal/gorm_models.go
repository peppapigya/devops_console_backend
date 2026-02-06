package dal

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

// HelmRepo Helm仓库模型
type HelmRepo struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null;column:name;size:255" json:"name"`
	URL       string    `gorm:"not null;column:url;size:500" json:"url"`
	Username  string    `gorm:"column:username;size:255" json:"username"`
	Password  string    `gorm:"column:password;size:500" json:"password"` // 加密存储
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (HelmRepo) TableName() string {
	return "helm_repo"
}

// HelmChart Helm应用缓存模型
type HelmChart struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RepoID      uint      `gorm:"index;not null;column:repo_id" json:"repo_id"`
	Name        string    `gorm:"index;not null;column:name;size:255" json:"name"`
	Version     string    `gorm:"not null;column:version;size:100" json:"version"`
	AppVersion  string    `gorm:"column:app_version;size:100" json:"app_version"`
	Description string    `gorm:"type:text;column:description" json:"description"`
	Icon        string    `gorm:"column:icon;size:500" json:"icon"`
	ChartURL    string    `gorm:"column:chart_url;size:500" json:"chart_url"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (HelmChart) TableName() string {
	return "helm_chart"
}

// HelmRelease Helm已安装应用实例模型
type HelmRelease struct {
	ID           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	InstanceID   uint      `gorm:"index;not null;column:instance_id" json:"instance_id"`
	Namespace    string    `gorm:"index;not null;column:namespace;size:255" json:"namespace"`
	ReleaseName  string    `gorm:"not null;column:release_name;size:255" json:"release_name"`
	ChartName    string    `gorm:"column:chart_name;size:255" json:"chart_name"`
	ChartVersion string    `gorm:"column:chart_version;size:100" json:"chart_version"`
	Status       string    `gorm:"column:status;size:50" json:"status"`
	Values       string    `gorm:"type:text;column:values" json:"values"` // JSON格式
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (HelmRelease) TableName() string {
	return "helm_release"
}
