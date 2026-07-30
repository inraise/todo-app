package statistics_postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/inraise/todo-app/internal/core/domain"
)

func (r *StatisticsRepository) GetTasks(
	ctx context.Context,
	userID *int,
	from *time.Time,
	to *time.Time,
) ([]domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, title, description, completed, created_at, completed_at, author_user_id
		FROM todoapp.tasks
		WHERE ($1::int IS NULL OR author_user_id = $1::int)
		AND ($2::timestamptz IS NULL OR created_at >= $2::timestamptz)
		AND ($3::timestamptz IS NULL OR created_at < $3::timestamptz)
		ORDER BY id ASC;`

	rows, err := r.pool.Query(
		ctx,
		query,
		userID,
		from,
		to,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to execute query to get tasks: %w",
			err,
		)
	}
	defer rows.Close()

	var taskModels []TaskModel
	for rows.Next() {
		var taskModel TaskModel
		err := rows.Scan(
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
			return nil, fmt.Errorf("scan tasks: %w", err)
		}

		taskModels = append(taskModels, taskModel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	taskDomains := taskDomainsFromModels(taskModels)

	return taskDomains, nil
}
