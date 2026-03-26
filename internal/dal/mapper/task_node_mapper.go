package mapper

import (
	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
)

// TaskNodeMapper 任务节点数据访问
type TaskNodeMapper struct {
	DB *gorm.DB
}

// NewTaskNodeMapper 创建任务节点数据访问实例
func NewTaskNodeMapper(db *gorm.DB) *TaskNodeMapper {
	return &TaskNodeMapper{DB: db}
}

// Create 创建节点
func (m *TaskNodeMapper) Create(node *model.TaskNode) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(node).Error
	})
}

// GetByID 根据 ID 查询
func (m *TaskNodeMapper) GetByID(id uint64) (*model.TaskNode, error) {
	var node model.TaskNode
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// Update 更新节点
func (m *TaskNodeMapper) Update(node *model.TaskNode) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Save(node).Error
	})
}

// Delete 删除节点
func (m *TaskNodeMapper) Delete(id uint64) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).Delete(&model.TaskNode{}).Error
	})
}

// DeleteByWorkflowID 删除指定工作流的所有节点
func (m *TaskNodeMapper) DeleteByWorkflowID(workflowID uint64) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Where("workflow_id = ?", workflowID).Delete(&model.TaskNode{}).Error
	})
}

// ListByWorkflowID 根据工作流ID查询所有节点
func (m *TaskNodeMapper) ListByWorkflowID(workflowID uint64) ([]*model.TaskNode, error) {
	var nodes []*model.TaskNode
	err := m.DB.Where("workflow_id = ? AND deleted_at IS NULL", workflowID).
		Order("sort_order ASC, id ASC").
		Find(&nodes).Error
	return nodes, err
}
