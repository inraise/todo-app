package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	coreLogger "github.com/inraise/todo-app/internal/core/logger"
	"github.com/inraise/todo-app/internal/core/repository/postgres/pool/pgx"
	"github.com/inraise/todo-app/internal/core/transport/http/middleware"
	"github.com/inraise/todo-app/internal/core/transport/http/server"
	postgres_repository "github.com/inraise/todo-app/internal/features/users/repository/postgres"
	"github.com/inraise/todo-app/internal/features/users/service"
	"github.com/inraise/todo-app/internal/features/users/transport/http"
	"go.uber.org/zap"
)

func main() {
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
	httpServer.RegisterAPIRoutes(apiVersionRouter)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
