package dal

import (
	"database/sql"
	"time"
)

// CustomMonitor 自定义监控配置模型
type CustomMonitor struct {
	ID             uint         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	AccountID      uint         `gorm:"column:account_id;not null;index:idx_custom_monitor_account" json:"account_id"`  // 所属用户
	TargetType     string       `gorm:"column:target_type;not null;index:idx_custom_monitor_target" json:"target_type"` // 'pod' 或 'node'
	Title          string       `gorm:"column:title;not null" json:"title"`                                             // 图表标题
	PromQLTemplate string       `gorm:"column:promql_template;type:text;not null" json:"promql_template"`               // PromQL 模板
	ChartType      string       `gorm:"column:chart_type;not null;default:'line'" json:"chart_type"`                    // 'line', 'bar' 等
	UnitSuffix     string       `gorm:"column:unit_suffix;default:''" json:"unit_suffix"`                               // Y轴单位后缀
	ColorTheme     string       `gorm:"column:color_theme;default:''" json:"color_theme"`                               // 颜色主题
	CreatedAt      *time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time   `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      sql.NullTime `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// TableName 指定表名
func (CustomMonitor) TableName() string {
	return "custom_monitors"
}
