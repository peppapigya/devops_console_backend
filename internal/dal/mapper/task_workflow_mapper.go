package mapper

import (
	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
)

type TaskWorkflowMapper struct {
	DB *gorm.DB
}

func NewTaskWorkflowMapper(db *gorm.DB) *TaskWorkflowMapper {
	return &TaskWorkflowMapper{DB: db}
}

func (m *TaskWorkflowMapper) Create(workflow *model.TaskWorkflow) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(workflow).Error
	})
}

func (m *TaskWorkflowMapper) GetByID(id uint64) (*model.TaskWorkflow, error) {
	var workflow model.TaskWorkflow
	err := m.DB.Where("id = ?", id).First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (m *TaskWorkflowMapper) Update(workflow *model.TaskWorkflow) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Save(workflow).Error
	})
}

func (m *TaskWorkflowMapper) Delete(id uint64) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).Delete(&model.TaskWorkflow{}).Error
	})
}

func (m *TaskWorkflowMapper) ListPage(page, pageSize int, name string, status int) (total int64, list []*model.TaskWorkflow, err error) {
	db := m.DB.Model(&model.TaskWorkflow{})

	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return 0, nil, err
	}

	err = db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	if err != nil {
		return 0, nil, err
	}

	return total, list, nil
}
