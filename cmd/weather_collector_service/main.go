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
		services     = app.InitServices(ctx, provider, clients, repositories)
		_            = app.InitSchedulers(ctx, provider, services)
	)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-signalChan

	logger.Info(ctx, "server gracefully shutdowned")
}
