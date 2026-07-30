package statistics_service

import (
	"context"
	"fmt"
	"time"

	"github.com/inraise/todo-app/internal/core/domain"
	core_errors "github.com/inraise/todo-app/internal/core/errors"
)

func (s *StatisticsService) GetStatistics(
	ctx context.Context,
	userID *int,
	from *time.Time,
	to *time.Time,
) (domain.Statistics, error) {
	if from != nil && to != nil {
		if to.Before(*from) || to.Equal(*from) {
			return domain.Statistics{}, fmt.Errorf(
				"`to` must be after `from`: %w",
				core_errors.ErrInvalidArgument,
			)
		}
	}

	tasks, err := s.statisticsRepository.GetTasks(ctx, userID, from, to)
	if err != nil {
		return domain.Statistics{}, fmt.Errorf(
			"failed to get tasks from repository: %w",
			err,
		)
	}

	statistics := calculateStatistics(tasks)

	return statistics, nil
}

func calculateStatistics(tasks []domain.Task) domain.Statistics {
	if len(tasks) == 0 {
		return domain.NewStatistics(0, 0, nil, nil)
	}

	tasksCreated := len(tasks)
	tasksCompleted := 0
	var totalCompletionTime time.Duration

	for _, task := range tasks {
		if task.Completed {
			tasksCompleted++
		}

		completionDuration := task.CompletionDuration()
		if completionDuration != nil {
			totalCompletionTime += *completionDuration
		}
	}

	tasksCompletedRate := float64(tasksCompleted) / float64(tasksCreated) * 100

	var averageCompletionTime *time.Duration
	if tasksCompleted > 0 && totalCompletionTime > 0 {
		avg := totalCompletionTime / time.Duration(tasksCompleted)
		averageCompletionTime = &avg
	}

	return domain.NewStatistics(
		tasksCreated,
		tasksCompleted,
		&tasksCompletedRate,
		averageCompletionTime,
	)
}
