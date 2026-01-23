package models

import (
	"database/sql"
	"time"
)

// Account 用户账号模型
type Account struct {
	ID        uint         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`                                  // 主键id
	Username  string       `gorm:"column:username;not null;default:'';index:idx_account_user_id" json:"username"` // 用户名
	Password  string       `gorm:"column:password;not null;default:''" json:"password"`                           // 密码
	Status    uint8        `gorm:"column:status;not null" json:"status"`                                          // 状态，0可用，1不可用
	Nickname  string       `gorm:"column:nickname;default:''" json:"nickname"`                                    // 昵称
	CreatedAt *time.Time   `gorm:"column:created_at" json:"created_at"`                                           // 创建时间
	UpdatedAt *time.Time   `gorm:"column:updated_at" json:"updated_at"`                                           // 更新时间
	DeletedAt sql.NullTime `gorm:"column:deleted_at;index" json:"deleted_at"`                                     // 删除时间
}

// TableName 指定表名
func (Account) TableName() string {
	return "account"
}
