package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameTaskTemplate = "task_templates"

// TaskTemplate 任务模板
type TaskTemplate struct {
	ID          uint32         `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	Name        string         `gorm:"column:name;type:varchar(128);not null;comment:模板名称" json:"name"`
	Description *string        `gorm:"column:description;type:varchar(255);comment:模板描述" json:"description"`
	Category    string         `gorm:"column:category;type:varchar(32);not null;comment:分类" json:"category"`
	TaskType    string         `gorm:"column:task_type;type:varchar(32);not null;comment:任务类型" json:"taskType"`
	Icon        *string        `gorm:"column:icon;type:varchar(255);comment:图标" json:"icon"`
	Config      *string        `gorm:"column:config;type:json;comment:默认配置(JSON)" json:"config"`
	Variables   *string        `gorm:"column:variables;type:json;comment:变量定义(JSON)" json:"variables"`
	IsSystem    bool           `gorm:"column:is_system;type:tinyint(1);not null;default:0;comment:是否系统预置" json:"isSystem"`
	IsPublic    bool           `gorm:"column:is_public;type:tinyint(1);not null;default:1;comment:是否公开" json:"isPublic"`
	CreatedBy   uint32         `gorm:"column:created_by;type:int unsigned;not null;comment:创建人ID" json:"createdBy"`
	UsageCount  int32          `gorm:"column:usage_count;type:int;not null;default:0;comment:使用次数" json:"usageCount"`
	CreatedAt   *time.Time     `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:datetime" json:"deletedAt"`
}

// TableName TaskTemplate's table name
func (*TaskTemplate) TableName() string {
	return TableNameTaskTemplate
}
