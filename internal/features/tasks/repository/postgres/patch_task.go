package task_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/inraise/todo-app/internal/core/domain"
	core_errors "github.com/inraise/todo-app/internal/core/errors"
	"github.com/inraise/todo-app/internal/core/repository/postgres/pool"
)

func (r *TasksRepository) PatchTask(
	ctx context.Context,
	taskID int,
	task domain.Task,
) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE todoapp.tasks
		SET title = $1, description = $2, completed = $3, completed_at = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING id, version, title, description, completed, created_at, completed_at, author_user_id;`

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Completed,
		task.CompletedAt,
		taskID,
		task.Version,
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
				"task with ID %d concurrently accessed: %w",
				taskID,
				core_errors.ErrNotFound,
			)
		}

		return domain.Task{}, fmt.Errorf(
			"failed to scan row into TaskModel: %w",
			err,
		)
	}

	taskDomain := taskDomainFromModel(taskModel)

	return taskDomain, nil
}
