package model

import (
	"time"

	"gorm.io/gorm"
)

// ===================== 域名监控 =====================

const TableNameMonitorDomain = "monitor_domains"

type MonitorDomain struct {
	ID            uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Domain        string         `gorm:"column:domain;type:varchar(255);not null;uniqueIndex" json:"domain"`
	Tags          StringSlice    `gorm:"column:tags;type:json" json:"tags"`
	Protocol      string         `gorm:"column:protocol;type:varchar(10);not null;default:https" json:"protocol"`
	CheckInterval int            `gorm:"column:check_interval;not null;default:300" json:"checkInterval"`
	Enabled       bool           `gorm:"column:enabled;not null;default:1" json:"enabled"`
	Status        string         `gorm:"column:status;type:varchar(20);default:unknown" json:"status"`
	StatusCode    *int           `gorm:"column:status_code" json:"statusCode"`
	ResponseTime  *int           `gorm:"column:response_time" json:"responseTime"`
	SslExpiry     *time.Time     `gorm:"column:ssl_expiry;type:datetime(3)" json:"sslExpiry"`
	SslDaysLeft   *int           `gorm:"column:ssl_days_left" json:"sslDaysLeft"`
	CertProvider  *string        `gorm:"column:cert_provider;type:varchar(100)" json:"certProvider"`
	LastCheck     *time.Time     `gorm:"column:last_check;type:datetime(3)" json:"lastCheck"`
	Remark        *string        `gorm:"column:remark;type:text" json:"remark"`
	CreatedAt     *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt     *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"-"`
}

func (*MonitorDomain) TableName() string { return TableNameMonitorDomain }

// ===================== SSL 证书 =====================

const TableNameMonitorSslCert = "monitor_ssl_certs"

type MonitorSslCert struct {
	ID           uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Domain       string         `gorm:"column:domain;type:varchar(255);not null;index" json:"domain"`
	DnsConfigID  *uint64        `gorm:"column:dns_config_id;type:bigint unsigned" json:"dnsConfigId"`
	DnsConfig    string         `gorm:"column:dns_config;type:varchar(100)" json:"dnsConfig"` // 冗余名称
	CertSource   string         `gorm:"column:cert_source;type:varchar(20);not null;default:ACME" json:"certSource"`
	CAProvider   string         `gorm:"column:ca_provider;type:varchar(50)" json:"caProvider"`
	KeyAlgorithm string         `gorm:"column:key_algorithm;type:varchar(20);not null;default:EC256" json:"keyAlgorithm"`
	Email        string         `gorm:"column:email;type:varchar(200)" json:"email"`
	Status       int8           `gorm:"column:status;not null;default:-1" json:"status"`        // -1申请中 1已签发 0失败 2过期
	InstanceID   *string        `gorm:"column:instance_id;type:varchar(100)" json:"instanceId"` // 云厂商免费证书的实例ID
	CertID       *string        `gorm:"column:cert_id;type:varchar(100)" json:"certId"`         // 第三方证书ID（如阿里云的certId）
	CertPem      *string        `gorm:"column:cert_pem;type:longtext" json:"-"`
	KeyPem       *string        `gorm:"column:key_pem;type:longtext" json:"-"`
	ExpireAt     *time.Time     `gorm:"column:expire_at;type:datetime(3)" json:"expireAt"`
	CreatedAt    *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt    *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"-"`
}

func (*MonitorSslCert) TableName() string { return TableNameMonitorSslCert }

// ===================== DNS 云厂商配置 =====================

const TableNameMonitorDnsProvider = "monitor_dns_providers"

type MonitorDnsProvider struct {
	ID           uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Name         string         `gorm:"column:name;type:varchar(200);not null" json:"name"`
	Provider     string         `gorm:"column:provider;type:varchar(50);not null" json:"provider"`
	AccessKey    string         `gorm:"column:access_key;type:varchar(500);not null" json:"accessKey"`
	AccessSecret string         `gorm:"column:access_secret;type:varchar(1000);not null" json:"-"`
	ZoneID       *string        `gorm:"column:zone_id;type:varchar(200)" json:"zoneId"`
	Region       *string        `gorm:"column:region;type:varchar(50)" json:"region"`
	Email        string         `gorm:"column:email;type:varchar(200)" json:"email"`
	Phone        string         `gorm:"column:phone;type:varchar(30)" json:"phone"`
	Status       string         `gorm:"column:status;type:varchar(20);not null;default:active" json:"status"`
	CreatedAt    *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt    *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"-"`
}

func (*MonitorDnsProvider) TableName() string { return TableNameMonitorDnsProvider }

// ===================== 故障记录 =====================

const TableNameMonitorIncident = "monitor_incidents"

type MonitorIncident struct {
	ID           uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	AlertTime    time.Time      `gorm:"column:alert_time;type:datetime(3);not null" json:"alertTime"`
	BusinessLine string         `gorm:"column:business_line;type:varchar(100);not null" json:"businessLine"`
	Level        string         `gorm:"column:level;type:varchar(10);not null;default:P4" json:"level"`
	Frequency    string         `gorm:"column:frequency;type:varchar(20);not null;default:偶发" json:"frequency"`
	AlertDesc    string         `gorm:"column:alert_desc;type:varchar(500);not null" json:"alertDesc"`
	Detail       *string        `gorm:"column:detail;type:text" json:"detail"`
	Dept         string         `gorm:"column:dept;type:varchar(100)" json:"dept"`
	Handler      string         `gorm:"column:handler;type:varchar(100);not null;default:admin" json:"handler"`
	Status       string         `gorm:"column:status;type:varchar(20);not null;default:pending" json:"status"`
	ResolvedAt   *time.Time     `gorm:"column:resolved_at;type:datetime(3)" json:"resolvedAt"`
	CreatedAt    *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt    *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"-"`
}

func (*MonitorIncident) TableName() string { return TableNameMonitorIncident }
