package mapper

import (
	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
)

type TaskTemplateMapper struct {
	DB *gorm.DB
}

func NewTaskTemplateMapper(db *gorm.DB) *TaskTemplateMapper {
	return &TaskTemplateMapper{DB: db}
}

func (m *TaskTemplateMapper) Create(template *model.TaskTemplate) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(template).Error
	})
}

func (m *TaskTemplateMapper) GetByID(id uint64) (*model.TaskTemplate, error) {
	var template model.TaskTemplate
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (m *TaskTemplateMapper) Update(template *model.TaskTemplate) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Save(template).Error
	})
}

func (m *TaskTemplateMapper) Delete(id uint64) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).Delete(&model.TaskTemplate{}).Error
	})
}

func (m *TaskTemplateMapper) ListPage(page, pageSize int, category string, taskType string, isSystem int) (total int64, list []*model.TaskTemplate, err error) {
	db := m.DB.Model(&model.TaskTemplate{}).Where("deleted_at IS NULL")

	if category != "" {
		db = db.Where("category = ?", category)
	}
	if taskType != "" {
		db = db.Where("task_type = ?", taskType)
	}
	if isSystem != -1 {
		db = db.Where("is_system = ?", isSystem)
	}

	if err = db.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	if err = db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return 0, nil, err
	}

	return total, list, nil
}

func (m *TaskTemplateMapper) ListByCategory(category string) ([]*model.TaskTemplate, error) {
	var list []*model.TaskTemplate
	err := m.DB.Where("category = ? AND deleted_at IS NULL", category).Find(&list).Error
	return list, err
}

func (m *TaskTemplateMapper) IncrementUsageCount(id uint64) error {
	return m.DB.Model(&model.TaskTemplate{}).Where("id = ?", id).UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}
