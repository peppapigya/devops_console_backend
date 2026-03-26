package task_scheduler

import (
	"context"
	"devops-console-backend/internal/dal/model"
	"fmt"
)

type TaskExecutor interface {
	Execute(ctx context.Context, node *model.TaskNode) error
}

type DefaultTaskExecutor struct{}

func NewTaskExecutor() *DefaultTaskExecutor {
	return &DefaultTaskExecutor{}
}

func (e *DefaultTaskExecutor) Execute(ctx context.Context, node *model.TaskNode) error {
	switch node.TargetType {
	case "host":
		return e.executeHost(ctx, node)
	case "k8s":
		return e.executeK8s(ctx, node)
	case "db":
		return e.executeDB(ctx, node)
	default:
		return e.executeLocal(ctx, node)
	}
}

func (e *DefaultTaskExecutor) executeHost(ctx context.Context, node *model.TaskNode) error {
	fmt.Printf("Executing task %s on host %d\n", node.NodeName, node.TargetID)
	return nil
}

func (e *DefaultTaskExecutor) executeK8s(ctx context.Context, node *model.TaskNode) error {
	fmt.Printf("Executing task %s on k8s cluster %d\n", node.NodeName, node.TargetID)
	return nil
}

func (e *DefaultTaskExecutor) executeDB(ctx context.Context, node *model.TaskNode) error {
	fmt.Printf("Executing task %s on db instance %d\n", node.NodeName, node.TargetID)
	return nil
}

func (e *DefaultTaskExecutor) executeLocal(ctx context.Context, node *model.TaskNode) error {
	fmt.Printf("Executing task %s locally\n", node.NodeName)
	return nil
}
