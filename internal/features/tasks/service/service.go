package task_service

import (
	"context"

	"github.com/inraise/todo-app/internal/core/domain"
)

type TasksService struct {
	tasksRepository TasksRepository
}

type TasksRepository interface {
	CreateTask(
		ctx context.Context,
		task domain.Task,
	) (domain.Task, error)

	GetTasks(
		ctx context.Context,
		userID *int,
		limit *int,
		offset *int,
	) ([]domain.Task, error)

	GetTask(
		ctx context.Context,
		taskID int,
	) (domain.Task, error)
}

func NewTasksService(tasksRepository TasksRepository) *TasksService {
	return &TasksService{
		tasksRepository: tasksRepository,
	}
}
