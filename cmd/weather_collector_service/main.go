package main

import (
	"context"
	"log/slog"

	"github.com/meteogo/logger/pkg/logger"
)

func main() {
	ctx := context.Background()
	logger.InitLogger(logger.EnvTypeLocal, slog.LevelDebug)

	logger.Info(ctx, "logger initialized successfully")
}
