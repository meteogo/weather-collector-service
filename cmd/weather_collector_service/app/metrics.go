package app

import (
	"context"

	"github.com/meteogo/weather-collector-service/internal/metrics"
)

type Metrics struct {
	manager *metrics.Manager
}

func InitMetrics(ctx context.Context) Metrics {
	manager := metrics.NewManager()
	manager.Start(ctx)

	return Metrics{
		manager: manager,
	}
}
