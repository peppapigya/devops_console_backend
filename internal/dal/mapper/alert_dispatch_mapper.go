package mapper

import (
	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
)

type AlertDispatchMapper struct {
	DB *gorm.DB
}

func NewAlertDispatchMapper(db *gorm.DB) *AlertDispatchMapper {
	return &AlertDispatchMapper{DB: db}
}

func (m *AlertDispatchMapper) Create(dispatch *model.AlertDispatch) error {
	return m.DB.Create(dispatch).Error
}

func (m *AlertDispatchMapper) GetByFingerprint(fingerprint string) (*model.AlertDispatch, error) {
	var dispatch model.AlertDispatch
	err := m.DB.Where("fingerprint = ?", fingerprint).First(&dispatch).Error
	return &dispatch, err
}

func (m *AlertDispatchMapper) UpdateFields(id uint64, fields map[string]interface{}) error {
	return m.DB.Model(&model.AlertDispatch{}).Where("id = ?", id).Updates(fields).Error
}
