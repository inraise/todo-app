package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/inraise/todo-app/internal/core/config"
	coreLogger "github.com/inraise/todo-app/internal/core/logger"
	"github.com/inraise/todo-app/internal/core/repository/postgres/pool/pgx"
	"github.com/inraise/todo-app/internal/core/transport/http/middleware"
	"github.com/inraise/todo-app/internal/core/transport/http/server"
	statistics_postgres "github.com/inraise/todo-app/internal/features/statistics/repository/postgres"
	statistics_service "github.com/inraise/todo-app/internal/features/statistics/service"
	statistics_http "github.com/inraise/todo-app/internal/features/statistics/transport/http"
	task_postgres_repository "github.com/inraise/todo-app/internal/features/tasks/repository/postgres"
	task_service "github.com/inraise/todo-app/internal/features/tasks/service"
	task_http "github.com/inraise/todo-app/internal/features/tasks/transport/http"
	postgres_repository "github.com/inraise/todo-app/internal/features/users/repository/postgres"
	"github.com/inraise/todo-app/internal/features/users/service"
	"github.com/inraise/todo-app/internal/features/users/transport/http"
	"go.uber.org/zap"
)

func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := coreLogger.NewLogger(coreLogger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init app logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("app time zone", zap.Any("time_zone", time.Local))
	logger.Debug("initializing postgres connection pool")
	pool, err := pgx.NewPool(
		ctx,
		pgx.NewConfigMust(),
	)
	if err != nil {
		logger.Fatal("failed to init connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := postgres_repository.NewUsersRepository(pool)
	usersService := service.NewUsersService(usersRepository)
	usersTransportHTTP := http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing feature", zap.String("feature", "tasks"))
	tasksRepository := task_postgres_repository.NewTasksRepository(pool)
	tasksService := task_service.NewTasksService(tasksRepository)
	tasksTransportHTTP := task_http.NewTasksHTTPHandler(tasksService)

	logger.Debug("initializing statistics", zap.String("feature", "statistics"))
	statisticsRepository := statistics_postgres.NewStatisticsRepository(pool)
	statisticsService := statistics_service.NewStatisticsService(statisticsRepository)
	statisticsTransportHTTP := statistics_http.NewStatisticsHTTPHandler(statisticsService)

	logger.Debug("initializing HTTP server")

	httpServer := server.NewHTTPServer(
		server.NewConfigMust(),
		logger,
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.Trace(),
		middleware.Panic(),
	)
	apiVersionRouter := server.NewAPIVersionRouter(server.ApiVersion1)
	apiVersionRouter.RegisterRoutes(usersTransportHTTP.Routes()...)
	apiVersionRouter.RegisterRoutes(tasksTransportHTTP.Routes()...)
	apiVersionRouter.RegisterRoutes(statisticsTransportHTTP.Routes()...)
	httpServer.RegisterAPIRoutes(apiVersionRouter)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
