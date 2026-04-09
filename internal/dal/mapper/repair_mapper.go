package mapper

import (
	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
)

// ==================== RepairSessionMapper ====================

type RepairSessionMapper struct {
	DB *gorm.DB
}

func NewRepairSessionMapper(db *gorm.DB) *RepairSessionMapper {
	return &RepairSessionMapper{DB: db}
}

func (m *RepairSessionMapper) Create(s *model.RepairSession) error {
	return m.DB.Create(s).Error
}

func (m *RepairSessionMapper) GetByID(id string) (*model.RepairSession, error) {
	var s model.RepairSession
	err := m.DB.Where("id = ?", id).First(&s).Error
	return &s, err
}

func (m *RepairSessionMapper) UpdateStatus(id, status string) error {
	return m.DB.Model(&model.RepairSession{}).Where("id = ?", id).
		Update("status", status).Error
}

func (m *RepairSessionMapper) UpdateFields(id string, fields map[string]interface{}) error {
	return m.DB.Model(&model.RepairSession{}).Where("id = ?", id).Updates(fields).Error
}

func (m *RepairSessionMapper) ListPage(page, pageSize int) (int64, []model.RepairSession, error) {
	var total int64
	var list []model.RepairSession
	db := m.DB.Model(&model.RepairSession{})
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	return total, list, err
}

// ==================== SessionMessageMapper ====================

type SessionMessageMapper struct {
	DB *gorm.DB
}

func NewSessionMessageMapper(db *gorm.DB) *SessionMessageMapper {
	return &SessionMessageMapper{DB: db}
}

func (m *SessionMessageMapper) BatchCreate(msgs []*model.SessionMessage) error {
	if len(msgs) == 0 {
		return nil
	}
	return m.DB.Create(&msgs).Error
}

// GetBySessionID 获取 session 的所有消息，摘要行合并到正常历史中
func (m *SessionMessageMapper) GetBySessionID(sessionID string) ([]model.SessionMessage, error) {
	var msgs []model.SessionMessage
	err := m.DB.Where("session_id = ?", sessionID).Order("id ASC").Find(&msgs).Error
	return msgs, err
}

func (m *SessionMessageMapper) CountBySessionID(sessionID string) (int64, error) {
	var count int64
	err := m.DB.Model(&model.SessionMessage{}).Where("session_id = ? AND is_summary = false", sessionID).Count(&count).Error
	return count, err
}

// DeleteOldNonSummary 删除 [startID, endID] 范围内的非摘要消息（压缩后替换用）
func (m *SessionMessageMapper) DeleteOldNonSummary(sessionID string, endID uint64) error {
	return m.DB.Where("session_id = ? AND id <= ? AND is_summary = false", sessionID, endID).
		Delete(&model.SessionMessage{}).Error
}

// ==================== RepairActionMapper ====================

type RepairActionMapper struct {
	DB *gorm.DB
}

func NewRepairActionMapper(db *gorm.DB) *RepairActionMapper {
	return &RepairActionMapper{DB: db}
}

func (m *RepairActionMapper) BatchCreate(actions []*model.RepairAction) error {
	if len(actions) == 0 {
		return nil
	}
	return m.DB.Create(&actions).Error
}

func (m *RepairActionMapper) GetBySessionID(sessionID string) ([]model.RepairAction, error) {
	var actions []model.RepairAction
	err := m.DB.Where("session_id = ?", sessionID).Order("action_order ASC").Find(&actions).Error
	return actions, err
}

func (m *RepairActionMapper) GetByID(id uint64) (*model.RepairAction, error) {
	var a model.RepairAction
	err := m.DB.Where("id = ?", id).First(&a).Error
	return &a, err
}

func (m *RepairActionMapper) UpdateFields(id uint64, fields map[string]interface{}) error {
	return m.DB.Model(&model.RepairAction{}).Where("id = ?", id).Updates(fields).Error
}
