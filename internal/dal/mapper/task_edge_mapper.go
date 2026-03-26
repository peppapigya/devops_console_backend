package mapper

import (
	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
)

type TaskEdgeMapper struct {
	DB *gorm.DB
}

func NewTaskEdgeMapper(db *gorm.DB) *TaskEdgeMapper {
	return &TaskEdgeMapper{DB: db}
}

func (m *TaskEdgeMapper) BatchCreate(edges []*model.TaskEdge) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(edges, 100).Error
	})
}

func (m *TaskEdgeMapper) GetByID(id uint64) (*model.TaskEdge, error) {
	var edge model.TaskEdge
	err := m.DB.Where("id = ?", id).First(&edge).Error
	if err != nil {
		return nil, err
	}
	return &edge, nil
}

func (m *TaskEdgeMapper) Delete(id uint64) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Unscoped().Where("id = ?", id).Delete(&model.TaskEdge{}).Error
	})
}

func (m *TaskEdgeMapper) DeleteByWorkflowID(workflowID uint64) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Unscoped().Where("workflow_id = ?", workflowID).Delete(&model.TaskEdge{}).Error
	})
}

func (m *TaskEdgeMapper) ListByWorkflowID(workflowID uint64) ([]*model.TaskEdge, error) {
	var edges []*model.TaskEdge
	err := m.DB.Where("workflow_id = ?", workflowID).Find(&edges).Error
	return edges, err
}

func (m *TaskEdgeMapper) ListBySourceNode(nodeID uint64) ([]*model.TaskEdge, error) {
	var edges []*model.TaskEdge
	err := m.DB.Where("source_node_id = ?", nodeID).Find(&edges).Error
	return edges, err
}

func (m *TaskEdgeMapper) ListByTargetNode(nodeID uint64) ([]*model.TaskEdge, error) {
	var edges []*model.TaskEdge
	err := m.DB.Where("target_node_id = ?", nodeID).Find(&edges).Error
	return edges, err
}
