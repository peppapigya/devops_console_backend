package scheduler

import (
	"context"
	"devops-console-backend/internal/dal/model"
	"sync"

	"github.com/robfig/cron/v3"
)

type WorkflowExecutor interface {
	ExecuteWorkflow(ctx context.Context, workflowID uint64, nodes []*model.TaskNode, edges []*model.TaskEdge, triggeredBy uint64) (uint64, error)
	ExecuteWorkflowWithTrigger(ctx context.Context, workflowID uint64, nodes []*model.TaskNode, edges []*model.TaskEdge, triggeredBy uint64, triggerType string) (uint64, error)
}

type CronScheduler struct {
	cron     *cron.Cron
	entries  map[uint64]cron.EntryID
	mu       sync.RWMutex
	executor WorkflowExecutor
}

var scheduler *CronScheduler
var once sync.Once

func GetScheduler(executor WorkflowExecutor) *CronScheduler {
	once.Do(func() {
		scheduler = &CronScheduler{
			cron:    cron.New(cron.WithSeconds()),
			entries: make(map[uint64]cron.EntryID),
		}
		scheduler.cron.Start()
	})

	if executor != nil && scheduler.executor == nil {
		scheduler.mu.Lock()
		scheduler.executor = executor
		scheduler.mu.Unlock()
	}

	return scheduler
}

func (s *CronScheduler) AddWorkflow(workflow *model.TaskWorkflow, nodes []*model.TaskNode, edges []*model.TaskEdge) error {
	if workflow.CronExpression == nil || *workflow.CronExpression == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.entries[uint64(workflow.ID)]; exists {
		s.cron.Remove(entryID)
		delete(s.entries, uint64(workflow.ID))
	}

	entryID, err := s.cron.AddFunc(*workflow.CronExpression, func() {
		if s.executor != nil {
			s.executor.ExecuteWorkflowWithTrigger(context.Background(), uint64(workflow.ID), nodes, edges, 0, "schedule")
		}
	})

	if err != nil {
		return err
	}

	s.entries[uint64(workflow.ID)] = entryID
	return nil
}

func (s *CronScheduler) RemoveWorkflow(workflowID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.entries[workflowID]; exists {
		s.cron.Remove(entryID)
		delete(s.entries, workflowID)
	}
}

func (s *CronScheduler) UpdateWorkflow(workflow *model.TaskWorkflow, nodes []*model.TaskNode, edges []*model.TaskEdge) error {
	s.RemoveWorkflow(uint64(workflow.ID))

	if workflow.Status == 1 && workflow.CronExpression != nil && *workflow.CronExpression != "" {
		return s.AddWorkflow(workflow, nodes, edges)
	}

	return nil
}

func (s *CronScheduler) Stop() {
	s.cron.Stop()
}

func (s *CronScheduler) GetScheduledWorkflows() []uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]uint64, 0, len(s.entries))
	for id := range s.entries {
		ids = append(ids, id)
	}
	return ids
}
