package statistics_http

import (
	"context"
	"net/http"
	"time"

	"github.com/inraise/todo-app/internal/core/domain"
	"github.com/inraise/todo-app/internal/core/transport/http/server"
)

type StatisticsHTTPHandler struct {
	statsService StatisticsService
}

type StatisticsService interface {
	GetStatistics(
		ctx context.Context,
		userID *int,
		from *time.Time,
		to *time.Time,
	) (domain.Statistics, error)
}

func NewStatisticsHTTPHandler(statsService StatisticsService) *StatisticsHTTPHandler {
	return &StatisticsHTTPHandler{
		statsService: statsService,
	}
}

func (h *StatisticsHTTPHandler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/statistics",
			Handler: h.GetStatistics,
		},
	}
}
