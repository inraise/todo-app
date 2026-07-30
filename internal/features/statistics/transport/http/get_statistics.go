package statistics_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/inraise/todo-app/internal/core/domain"
	"github.com/inraise/todo-app/internal/core/logger"
	core_request "github.com/inraise/todo-app/internal/core/transport/http/request"
	"github.com/inraise/todo-app/internal/core/transport/http/response"
)

type GetStatisticsResponse struct {
	TasksCreated               int      `json:"tasks_created"`
	TasksCompleted             int      `json:"tasks_completed"`
	TasksCompletedRate         *float64 `json:"tasks_completed_rate"`
	TasksAverageCompletionTime *string  `json:"tasks_average_completion_time"`
}

func (h *StatisticsHTTPHandler) GetStatistics(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response.NewHTTPResponseHandler(log, rw)

	userID, from, to, err := getUserIdFromToQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get query params",
		)

		return
	}

	stats, err := h.statsService.GetStatistics(ctx, userID, from, to)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get statistics",
		)

		return
	}

	response := toDTOFromDomain(stats)
	responseHandler.JSONResponse(
		response,
		http.StatusOK,
	)
}

func toDTOFromDomain(stats domain.Statistics) GetStatisticsResponse {
	var avgTime *string
	if stats.TasksAverageCompletionTime != nil {
		avgTimeStr := stats.TasksAverageCompletionTime.String()
		avgTime = &avgTimeStr
	}

	return GetStatisticsResponse{
		TasksCreated:               stats.TasksCreated,
		TasksCompleted:             stats.TasksCompleted,
		TasksCompletedRate:         stats.TasksCompletedRate,
		TasksAverageCompletionTime: avgTime,
	}
}

func getUserIdFromToQueryParams(r *http.Request) (*int, *time.Time, *time.Time, error) {
	const (
		userIDQueryParamKey = "user_id"
		fromQueryParamKey   = "from"
		toQueryParamKey     = "to"
	)

	userID, err := core_request.GetIntQueryParam(r, userIDQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'user_id' query param: %w", err)
	}

	from, err := core_request.GetTimeQueryParam(r, fromQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'from' query param: %w", err)
	}

	to, err := core_request.GetTimeQueryParam(r, toQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get 'to' query param: %w", err)
	}

	return userID, from, to, nil
}
