package task_scheduler

import (
	"context"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/internal/dal/request"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/internal/services/scheduler"
	"encoding/json"
	"errors"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type TaskWorkflowService struct {
	workflowMapper  *mapper.TaskWorkflowMapper
	nodeMapper      *mapper.TaskNodeMapper
	edgeMapper      *mapper.TaskEdgeMapper
	executionMapper *mapper.TaskExecutionMapper
	nodeExecMapper  *mapper.TaskNodeExecutionMapper
	dagExecutor     *DAGExecutor
}

func validateCronExpression(cronExpr string) error {
	if cronExpr == "" {
		return nil // 允许空（手动触发）
	}
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(cronExpr)
	if err != nil {
		return errors.New("cron expression invalid: " + err.Error())
	}
	return nil
}

func NewTaskWorkflowService(
	workflowMapper *mapper.TaskWorkflowMapper,
	nodeMapper *mapper.TaskNodeMapper,
	edgeMapper *mapper.TaskEdgeMapper,
	executionMapper *mapper.TaskExecutionMapper,
	nodeExecMapper *mapper.TaskNodeExecutionMapper,
) *TaskWorkflowService {
	svc := &TaskWorkflowService{
		workflowMapper:  workflowMapper,
		nodeMapper:      nodeMapper,
		edgeMapper:      edgeMapper,
		executionMapper: executionMapper,
		nodeExecMapper:  nodeExecMapper,
	}
	svc.dagExecutor = NewDAGExecutor(executionMapper, nodeExecMapper)
	scheduler.GetScheduler(svc.dagExecutor)
	return svc
}

func (s *TaskWorkflowService) CreateWorkflow(req *request.TaskWorkflowCreateRequest, nodesReq []*request.TaskNodeCreateRequest, edgesReq []*request.TaskEdgeCreateRequest) (*model.TaskWorkflow, error) {
	if err := validateCronExpression(req.CronExpression); err != nil {
		return nil, err
	}
	var workflow model.TaskWorkflow
	err := s.workflowMapper.DB.Transaction(func(tx *gorm.DB) error {
		description := req.Description
		cronExpr := req.CronExpression
		now := time.Now()
		workflow = model.TaskWorkflow{
			Name:           req.Name,
			Description:    &description,
			CronExpression: &cronExpr,
			Status:         int32(req.Status),
			TaskType:       "workflow",
			CreatedAt:      &now,
			UpdatedAt:      &now,
		}
		if err := tx.Create(&workflow).Error; err != nil {
			return err
		}

		workflowID := workflow.ID

		tempIDToRealID := make(map[string]uint32)
		for _, nr := range nodesReq {
			config := nr.Config
			if config == "" {
				config = "{}"
			}
			node := &model.TaskNode{
				WorkflowID: workflowID,
				NodeName:   nr.Name,
				NodeType:   nr.Type,
				Config:     &config,
				PositionX:  float64(nr.PositionX),
				TargetID:   nr.TargetID,
				TargetType: nr.TargetType,
				PositionY:  float64(nr.PositionY),
				CreatedAt:  &now,
				UpdatedAt:  &now,
			}
			if err := tx.Create(node).Error; err != nil {
				return err
			}
			if nr.TempID != "" {
				tempIDToRealID[nr.TempID] = node.ID
			}
		}

		for _, er := range edgesReq {
			fromID := tempIDToRealID[er.SourceTempID]
			toID := tempIDToRealID[er.TargetTempID]

			condition := er.Condition
			edge := &model.TaskEdge{
				WorkflowID:   workflowID,
				SourceNodeID: fromID,
				TargetNodeID: toID,
				Condition:    &condition,
				EdgeType:     "default",
				SourceHandle: er.SourceHandle,
				TargetHandle: er.TargetHandle,
				CreatedAt:    &now,
			}
			if err := tx.Create(edge).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if workflow.CronExpression != nil && *workflow.CronExpression != "" && workflow.Status == 1 {
		nodes, _ := s.nodeMapper.ListByWorkflowID(uint64(workflow.ID))
		edges, _ := s.edgeMapper.ListByWorkflowID(uint64(workflow.ID))
		err := scheduler.GetScheduler(s.dagExecutor).AddWorkflow(&workflow, nodes, edges)
		if err != nil {
			return nil, err
		}
	}

	return &workflow, nil
}

func (s *TaskWorkflowService) GetWorkflowByID(id uint64) (*response.TaskWorkflowVO, error) {
	w, err := s.workflowMapper.GetByID(id)
	if err != nil {
		return nil, err
	}

	nodes, err := s.nodeMapper.ListByWorkflowID(id)
	if err != nil {
		return nil, err
	}

	edges, err := s.edgeMapper.ListByWorkflowID(id)
	if err != nil {
		return nil, err
	}

	vo := &response.TaskWorkflowVO{
		ID:        uint64(w.ID),
		Name:      w.Name,
		Status:    int(w.Status),
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
		Nodes:     make([]response.TaskNodeVO, 0, len(nodes)),
		Edges:     make([]response.TaskEdgeVO, 0, len(edges)),
	}
	if w.Description != nil {
		vo.Description = *w.Description
	}
	if w.CronExpression != nil {
		vo.CronExpression = *w.CronExpression
	}

	for _, n := range nodes {
		nVO := response.TaskNodeVO{
			ID:         uint64(n.ID),
			WorkflowID: uint64(n.WorkflowID),
			Name:       n.NodeName,
			Type:       n.NodeType,
			PositionX:  n.PositionX,
			PositionY:  n.PositionY,
			CreatedAt:  n.CreatedAt,
		}
		if n.Config != nil {
			nVO.Config = json.RawMessage(*n.Config)
		}
		vo.Nodes = append(vo.Nodes, nVO)
	}

	for _, e := range edges {
		eVO := response.TaskEdgeVO{
			ID:           uint64(e.ID),
			WorkflowID:   uint64(e.WorkflowID),
			FromNodeID:   uint64(e.SourceNodeID),
			ToNodeID:     uint64(e.TargetNodeID),
			SourceHandle: e.SourceHandle,
			TargetHandle: e.TargetHandle,
		}
		if e.Condition != nil {
			eVO.Condition = *e.Condition
		}
		vo.Edges = append(vo.Edges, eVO)
	}

	return vo, nil
}

func (s *TaskWorkflowService) UpdateWorkflow(id uint64, req *request.TaskWorkflowUpdateRequest, nodesReq []*request.TaskNodeCreateRequest, edgesReq []*request.TaskEdgeCreateRequest) error {
	if err := validateCronExpression(req.CronExpression); err != nil {
		return err
	}
	err := s.workflowMapper.DB.Transaction(func(tx *gorm.DB) error {
		w, err := s.workflowMapper.GetByID(id)
		if err != nil {
			return err
		}
		if req.Name != "" {
			w.Name = req.Name
		}
		if req.Description != "" {
			desc := req.Description
			w.Description = &desc
		}
		cronExpr := req.CronExpression
		w.CronExpression = &cronExpr
		w.Status = int32(req.Status)
		now := time.Now()
		w.UpdatedAt = &now

		if err := tx.Save(w).Error; err != nil {
			return err
		}

		if err := tx.Where("workflow_id = ?", id).Delete(&model.TaskNode{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workflow_id = ?", id).Delete(&model.TaskEdge{}).Error; err != nil {
			return err
		}

		tempIDToRealID := make(map[string]uint32)
		for _, nr := range nodesReq {
			config := nr.Config
			if config == "" {
				config = "{}"
			}
			node := &model.TaskNode{
				WorkflowID: uint32(id),
				NodeName:   nr.Name,
				NodeType:   nr.Type,
				Config:     &config,
				TargetID:   nr.TargetID,
				TargetType: nr.TargetType,
				PositionX:  float64(nr.PositionX),
				PositionY:  float64(nr.PositionY),
				UpdatedAt:  &now,
			}
			if err := tx.Create(node).Error; err != nil {
				return err
			}
			if nr.TempID != "" {
				tempIDToRealID[nr.TempID] = node.ID
			}
		}

		for _, er := range edgesReq {
			fromID := tempIDToRealID[er.SourceTempID]
			toID := tempIDToRealID[er.TargetTempID]

			condition := er.Condition
			edge := &model.TaskEdge{
				WorkflowID:   uint32(id),
				SourceNodeID: fromID,
				TargetNodeID: toID,
				Condition:    &condition,
				EdgeType:     "default",
				SourceHandle: er.SourceHandle,
				TargetHandle: er.TargetHandle,
				CreatedAt:    &now,
			}
			if err := tx.Create(edge).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	w, _ := s.workflowMapper.GetByID(id)
	nodes, _ := s.nodeMapper.ListByWorkflowID(id)
	edges, _ := s.edgeMapper.ListByWorkflowID(id)
	scheduler.GetScheduler(s.dagExecutor).UpdateWorkflow(w, nodes, edges)

	return nil
}

func (s *TaskWorkflowService) DeleteWorkflow(id uint64) error {
	err := s.workflowMapper.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&model.TaskWorkflow{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workflow_id = ?", id).Delete(&model.TaskNode{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workflow_id = ?", id).Delete(&model.TaskEdge{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	scheduler.GetScheduler(s.dagExecutor).RemoveWorkflow(id)
	return nil
}

func (s *TaskWorkflowService) ListWorkflows(page, pageSize int, name string, status int) (*response.TaskWorkflowListResponse, error) {
	total, list, err := s.workflowMapper.ListPage(page, pageSize, name, status)
	if err != nil {
		return nil, err
	}

	voList := make([]response.TaskWorkflowVO, 0, len(list))
	for _, w := range list {
		vo := response.TaskWorkflowVO{
			ID:        uint64(w.ID),
			Name:      w.Name,
			Status:    int(w.Status),
			CreatedAt: w.CreatedAt,
			UpdatedAt: w.UpdatedAt,
		}
		if w.Description != nil {
			vo.Description = *w.Description
		}
		if w.CronExpression != nil {
			vo.CronExpression = *w.CronExpression
		}
		voList = append(voList, vo)
	}

	return &response.TaskWorkflowListResponse{
		Total: total,
		List:  voList,
	}, nil
}

func (s *TaskWorkflowService) UpdateWorkflowStatus(id uint64, status int) error {
	w, err := s.workflowMapper.GetByID(id)
	if err != nil {
		return err
	}
	w.Status = int32(status)
	now := time.Now()
	w.UpdatedAt = &now
	if err := s.workflowMapper.Update(w); err != nil {
		return err
	}

	if w.CronExpression != nil && *w.CronExpression != "" {
		nodes, _ := s.nodeMapper.ListByWorkflowID(id)
		edges, _ := s.edgeMapper.ListByWorkflowID(id)
		scheduler.GetScheduler(s.dagExecutor).UpdateWorkflow(w, nodes, edges)
	}

	return nil
}

func (s *TaskWorkflowService) ExecuteWorkflow(workflowID uint64, triggeredBy uint64) (uint64, error) {
	nodes, err := s.nodeMapper.ListByWorkflowID(workflowID)
	if err != nil {
		return 0, err
	}

	edges, err := s.edgeMapper.ListByWorkflowID(workflowID)
	if err != nil {
		return 0, err
	}

	return s.dagExecutor.ExecuteWorkflow(context.Background(), workflowID, nodes, edges, triggeredBy)
}
