package mapper

import (
	"devops-console-backend/internal/dal/model"
	"encoding/json"

	"gorm.io/gorm"
)

type TaskNodeExecutionMapper struct {
	DB *gorm.DB
}

func NewTaskNodeExecutionMapper(db *gorm.DB) *TaskNodeExecutionMapper {
	return &TaskNodeExecutionMapper{DB: db}
}

func (m *TaskNodeExecutionMapper) Create(execution *model.TaskNodeExecution) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(execution).Error
	})
}

func (m *TaskNodeExecutionMapper) GetByID(id uint64) (*model.TaskNodeExecution, error) {
	var execution model.TaskNodeExecution
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&execution).Error
	return &execution, err
}

func (m *TaskNodeExecutionMapper) Update(execution *model.TaskNodeExecution) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Model(execution).Updates(execution).Error
	})
}

func (m *TaskNodeExecutionMapper) UpdateStatus(id uint64, status string, outputParams json.RawMessage, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if len(outputParams) > 0 {
		updates["output_params"] = string(outputParams)
	}
	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}
	return m.DB.Model(&model.TaskNodeExecution{}).Where("id = ?", id).Updates(updates).Error
}

func (m *TaskNodeExecutionMapper) UpdateLogs(id uint64, logs string) error {
	return m.DB.Model(&model.TaskNodeExecution{}).Where("id = ?", id).Update("logs", logs).Error
}

func (m *TaskNodeExecutionMapper) ListByExecutionID(executionID uint64) ([]*model.TaskNodeExecution, error) {
	var executions []*model.TaskNodeExecution
	err := m.DB.Where("execution_id = ? AND deleted_at IS NULL", executionID).Find(&executions).Error
	return executions, err
}

func (m *TaskNodeExecutionMapper) GetByExecutionAndNode(executionID, nodeID uint64) (*model.TaskNodeExecution, error) {
	var execution model.TaskNodeExecution
	err := m.DB.Where("execution_id = ? AND node_id = ? AND deleted_at IS NULL", executionID, nodeID).First(&execution).Error
	return &execution, err
}

func (m *TaskNodeExecutionMapper) BatchCreate(executions []*model.TaskNodeExecution) error {
	return m.DB.Create(executions).Error
}
