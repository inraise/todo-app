package task_service

import (
	"context"
	"fmt"

	"github.com/inraise/todo-app/internal/core/domain"
)

func (s *TasksService) PatchTask(
	ctx context.Context,
	taskID int,
	patch domain.TaskPatch,
) (domain.Task, error) {
	task, err := s.tasksRepository.GetTask(ctx, taskID)
	if err != nil {
		return domain.Task{}, fmt.Errorf("failed to get task with ID %d: %w", taskID, err)
	}

	if err := task.ApplyPatch(patch); err != nil {
		return domain.Task{}, fmt.Errorf(
			"failed to apply patch to task with ID %d: %w", 
			taskID, 
			err,
		)
	}

	patchedTask, err := s.tasksRepository.PatchTask(ctx, taskID, task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("failed to patch task with ID %d: %w", taskID, err)
	}

	return patchedTask, nil
}
