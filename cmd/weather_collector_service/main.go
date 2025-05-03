package main

import (
	"context"
	"log/slog"

	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/internal/services/weather_collector"
)

func main() {
	ctx := context.Background()
	logger.InitLogger(logger.EnvTypeLocal, slog.LevelDebug)
	logger.Info(ctx, "logger initialized successfully")

	provider := config.NewProvider(".cfg/values.yaml")
	logger.Info(ctx, "config provider created successfully")

	collectorConfig, err := weather_collector.NewConfig(provider)
	if err != nil {
		panic(err)
	}
	collectorService := weather_collector.NewService(collectorConfig, nil, nil)

	_ = collectorService
}
