package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// StringSlice JSON 数组类型，存储 tags 等
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
	return json.Unmarshal(b, s)
}

// ===================== 主机分组 =====================

const TableNameAssetHostGroup = "asset_host_groups"

type AssetHostGroup struct {
	ID        uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	ParentID  uint64         `gorm:"column:parent_id;type:bigint unsigned;not null;default:0" json:"parentId"`
	Name      string         `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Remark    *string        `gorm:"column:remark;type:text" json:"remark"`
	CreatedAt *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"-"`

	// 非数据库字段，构建树形
	Children  []*AssetHostGroup `gorm:"-" json:"children,omitempty"`
	HostCount int64             `gorm:"-" json:"hostCount"`
}

func (*AssetHostGroup) TableName() string { return TableNameAssetHostGroup }

// ===================== 主机 =====================

const TableNameAssetHost = "asset_hosts"

type AssetHost struct {
	ID            uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	GroupID       uint64         `gorm:"column:group_id;type:bigint unsigned;not null;default:0;index" json:"groupId"`
	Name          string         `gorm:"column:name;type:varchar(200);not null" json:"name"`
	IP            string         `gorm:"column:ip;type:varchar(50);not null" json:"ip"`
	Port          uint16         `gorm:"column:port;type:smallint unsigned;not null;default:22" json:"port"`
	OsType        string         `gorm:"column:os_type;type:varchar(20);not null;default:linux" json:"osType"`
	CloudProvider string         `gorm:"column:cloud_provider;type:varchar(30)" json:"cloudProvider"`
	Username      string         `gorm:"column:username;type:varchar(100);not null;default:root" json:"username"`
	AuthType      string         `gorm:"column:auth_type;type:varchar(20);not null;default:password" json:"authType"`
	Password      *string        `gorm:"column:password;type:varchar(500)" json:"-"`
	PrivateKey    *string        `gorm:"column:private_key;type:text" json:"-"`
	Tags          StringSlice    `gorm:"column:tags;type:json" json:"tags"`
	Status        string         `gorm:"column:status;type:varchar(20);not null;default:unknown" json:"status"`
	CPUUsage      *float64       `gorm:"column:cpu_usage;type:decimal(5,2)" json:"cpuUsage"`
	MemUsage      *float64       `gorm:"column:mem_usage;type:decimal(5,2)" json:"memUsage"`
	DiskUsage     *float64       `gorm:"column:disk_usage;type:decimal(5,2)" json:"diskUsage"`
	ProcessCount  *int           `gorm:"column:process_count" json:"processCount"`
	PortCount     *int           `gorm:"column:port_count" json:"portCount"`
	TunnelCount   *int           `gorm:"column:tunnel_count" json:"tunnelCount"`
	Remark        *string        `gorm:"column:remark;type:text" json:"remark"`
	CreatedAt     *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt     *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"-"`
}

func (*AssetHost) TableName() string { return TableNameAssetHost }
