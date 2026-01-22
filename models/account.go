package models

import (
	"time"
)

// Account 用户账号模型
type Account struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"` // 使用uint匹配UNSIGNED INT
	UserID    string    `gorm:"default:'';column:user_id" json:"user_id"`
	Password  string    `gorm:"default:'';column:password" json:"password"`
	Nickname  string    `gorm:"default:'';column:nickname" json:"nickname"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Account) TableName() string {
	return "account"
}
