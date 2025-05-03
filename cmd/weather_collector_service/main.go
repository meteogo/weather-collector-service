package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/logger/pkg/logger"
	appconfig "github.com/meteogo/weather-collector-service/internal/config"
)

func main() {
	ctx := context.Background()
	logger.InitLogger(logger.EnvTypeLocal, slog.LevelDebug)
	logger.Info(ctx, "logger initialized successfully")

	provider := config.NewProvider(".cfg/values.yaml")
	logger.Info(ctx, "config provider created successfully")
	logger.Debug(ctx, fmt.Sprintf("will collect weather every %v", provider.GetConfigClient().GetValue(appconfig.WeatherCollectorCronDuration).Duration()))
}
