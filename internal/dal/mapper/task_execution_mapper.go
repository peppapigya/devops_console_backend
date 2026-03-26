package mapper

import (
	"devops-console-backend/internal/dal/model"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// TaskExecutionMapper 任务执行记录数据访问
type TaskExecutionMapper struct {
	DB *gorm.DB
}

// NewTaskExecutionMapper 创建任务执行记录 Mapper
func NewTaskExecutionMapper(db *gorm.DB) *TaskExecutionMapper {
	return &TaskExecutionMapper{DB: db}
}

// Create 创建执行记录
func (m *TaskExecutionMapper) Create(execution *model.TaskExecution) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(execution).Error
	})
}

// GetByID 根据 ID 查询
func (m *TaskExecutionMapper) GetByID(id uint64) (*model.TaskExecution, error) {
	var execution model.TaskExecution
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&execution).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &execution, nil
}

// GetByExecutionNo 根据执行编号查询
func (m *TaskExecutionMapper) GetByExecutionNo(executionNo string) (*model.TaskExecution, error) {
	var execution model.TaskExecution
	err := m.DB.Where("execution_no = ? AND deleted_at IS NULL", executionNo).First(&execution).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &execution, nil
}

// Update 更新执行记录
func (m *TaskExecutionMapper) Update(execution *model.TaskExecution) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Model(execution).Updates(execution).Error
	})
}

// UpdateStatus 更新状态及结果
func (m *TaskExecutionMapper) UpdateStatus(id uint64, status string, result json.RawMessage, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if len(result) > 0 {
		resStr := string(result)
		updates["result"] = &resStr
	}
	if errorMsg != "" {
		updates["error_message"] = &errorMsg
	}

	// 如果状态是终态，则更新结束时间
	if status == "COMPLETED" || status == "FAILED" || status == "CANCELLED" || status == "success" || status == "failed" {
		now := time.Now()
		updates["end_time"] = &now
	}

	return m.DB.Model(&model.TaskExecution{}).Where("id = ?", id).Updates(updates).Error
}

// ListPage 分页查询
func (m *TaskExecutionMapper) ListPage(page, pageSize int, workflowID uint64, status string, triggerType string) (total int64, list []*model.TaskExecution, err error) {
	db := m.DB.Model(&model.TaskExecution{}).Where("deleted_at IS NULL")
	if workflowID > 0 {
		db = db.Where("workflow_id = ?", workflowID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if triggerType != "" {
		db = db.Where("trigger_type = ?", triggerType)
	}

	err = db.Count(&total).Error
	if err != nil {
		return 0, nil, err
	}

	err = db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return total, list, err
}

// ListByWorkflowID 根据工作流 ID 获取执行列表
func (m *TaskExecutionMapper) ListByWorkflowID(workflowID uint64, limit int) ([]*model.TaskExecution, error) {
	var list []*model.TaskExecution
	err := m.DB.Where("workflow_id = ? AND deleted_at IS NULL", workflowID).
		Order("id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// GetLatestByWorkflowID 获取最近一次执行
func (m *TaskExecutionMapper) GetLatestByWorkflowID(workflowID uint64) (*model.TaskExecution, error) {
	var execution model.TaskExecution
	err := m.DB.Where("workflow_id = ? AND deleted_at IS NULL", workflowID).
		Order("id DESC").
		First(&execution).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &execution, nil
}

// GetPendingRetries 获取待重试任务
func (m *TaskExecutionMapper) GetPendingRetries() ([]*model.TaskExecution, error) {
	var list []*model.TaskExecution
	now := time.Now()
	err := m.DB.Where("status = ? AND next_retry_time <= ? AND deleted_at IS NULL", "WAITING_RETRY", now).
		Find(&list).Error
	return list, err
}
