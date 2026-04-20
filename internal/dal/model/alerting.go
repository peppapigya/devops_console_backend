package model

import "time"

const TableNameAlertDispatch = "alert_dispatches"

type AlertDispatch struct {
	ID                 uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Fingerprint        string     `gorm:"column:fingerprint;type:varchar(191);not null;uniqueIndex" json:"fingerprint"`
	Status             string     `gorm:"column:status;type:varchar(32);not null;default:'received'" json:"status"`
	AlertStatus        string     `gorm:"column:alert_status;type:varchar(32);not null;default:'firing'" json:"alert_status"`
	AlertName          string     `gorm:"column:alert_name;type:varchar(191)" json:"alert_name"`
	AlertKey           string     `gorm:"column:alert_key;type:varchar(191);index" json:"alert_key"`
	SessionID          string     `gorm:"column:session_id;type:varchar(36);index" json:"session_id"`
	LogEventID         string     `gorm:"column:log_event_id;type:varchar(128);index" json:"log_event_id"`
	Host               string     `gorm:"column:host;type:varchar(255)" json:"host"`
	Service            string     `gorm:"column:service;type:varchar(255)" json:"service"`
	Level              string     `gorm:"column:level;type:varchar(32)" json:"level"`
	Message            string     `gorm:"column:message;type:text" json:"message"`
	RawPayload         string     `gorm:"column:raw_payload;type:mediumtext" json:"raw_payload"`
	FiringNotifiedAt   *time.Time `gorm:"column:firing_notified_at;type:datetime(3)" json:"firing_notified_at,omitempty"`
	ResultNotifiedAt   *time.Time `gorm:"column:result_notified_at;type:datetime(3)" json:"result_notified_at,omitempty"`
	ResolvedNotifiedAt *time.Time `gorm:"column:resolved_notified_at;type:datetime(3)" json:"resolved_notified_at,omitempty"`
	LastSessionStatus  string     `gorm:"column:last_session_status;type:varchar(32)" json:"last_session_status"`
	CreatedAt          *time.Time `gorm:"column:created_at;type:datetime(3)" json:"created_at"`
	UpdatedAt          *time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updated_at"`
}

func (*AlertDispatch) TableName() string { return TableNameAlertDispatch }
