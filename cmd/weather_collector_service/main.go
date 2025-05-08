package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/cmd/weather_collector_service/app"
	"github.com/meteogo/weather-collector-service/internal/closer"
)

func main() {
	var (
		env      = logger.EnvTypeLocal
		logLevel = slog.LevelDebug
	)

	ctx := context.Background()
	logger.InitLogger(logger.EnvType(env), logLevel)
	logger.Info(ctx, "logger initialized successfully", slog.Any("env", env), slog.Any("level", logLevel.String()))

	provider := config.NewProvider(".cfg/values.yaml")
	logger.Info(ctx, "config provider created successfully")

	var (
		repositories = app.InitRepositories(ctx, provider)
		clients      = app.InitClients()
		publishers   = app.InitPublishers(ctx, provider)
		metrics      = app.InitMetrics(ctx)
		services     = app.InitServices(provider, clients, publishers, repositories, metrics)
		_            = app.InitSchedulers(ctx, provider, services)
	)

	app.InitTracer(ctx, provider)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-signalChan

	if err := closer.Close(ctx); err != nil {
		logger.Error(ctx, "error while closer.Close()", slog.Any("error", err))
	}
	logger.Info(ctx, "server gracefully shutdowned")
}
