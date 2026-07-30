package task_service

import (
	"context"
	"fmt"

	"github.com/inraise/todo-app/internal/core/domain"
)

func (s *TasksService) GetTask(
	ctx context.Context,
	taskID int,
) (domain.Task, error) {
	task, err := s.tasksRepository.GetTask(ctx, taskID)
	if err != nil {
		return domain.Task{}, fmt.Errorf(
			"failed to get task with ID %d: %w",
			taskID,
			err,
		)
	}

	return task, nil
}
