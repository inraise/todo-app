package task_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/inraise/todo-app/internal/core/domain"
	core_errors "github.com/inraise/todo-app/internal/core/errors"
	"github.com/inraise/todo-app/internal/core/repository/postgres/pool"
)

func (r *TasksRepository) GetTask(
	ctx context.Context,
	taskID int,
) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, title, description, completed, created_at, completed_at, author_user_id
		FROM todoapp.tasks
		WHERE id = $1;`

	row := r.pool.QueryRow(
		ctx,
		query,
		taskID,
	)

	var taskModel TaskModel
	err := row.Scan(
		&taskModel.ID,
		&taskModel.Version,
		&taskModel.Title,
		&taskModel.Description,
		&taskModel.Completed,
		&taskModel.CreatedAt,
		&taskModel.CompletedAt,
		&taskModel.AuthorUserID,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.Task{}, fmt.Errorf(
				"task with ID %d does not exist: %w",
				taskID,
				core_errors.ErrNotFound,
			)
		}

		return domain.Task{}, fmt.Errorf(
			"failed to scan row: %w",
			err,
		)
	}

	taskDomain := domain.NewTask(
		taskModel.ID,
		taskModel.Version,
		taskModel.Title,
		taskModel.Description,
		taskModel.Completed,
		taskModel.CreatedAt,
		taskModel.CompletedAt,
		taskModel.AuthorUserID,
	)

	return taskDomain, nil
}
