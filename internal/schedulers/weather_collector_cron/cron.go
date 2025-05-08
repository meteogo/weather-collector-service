package weather_collector_cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/meteogo/logger/pkg/logger"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
)

type Config interface {
	Duration() time.Duration
}

type Service interface {
	CollectData(ctx context.Context) error
}

type Cron struct {
	config  Config
	cron    *cron.Cron
	service Service
}

func NewCron(config Config, cron *cron.Cron, service Service) *Cron {
	return &Cron{
		config:  config,
		cron:    cron,
		service: service,
	}
}

func (c *Cron) Start(ctx context.Context) {
	c.cron.Schedule(cron.Every(c.config.Duration()), cron.FuncJob(func() {
		c.Do(ctx)
	}))

	c.cron.Start()
	logger.Info(ctx, "weather collector cron successfully started", slog.String("duration", c.config.Duration().String()))
}

func (c *Cron) Do(ctx context.Context) {
	start := time.Now()
	defer func() {
		logger.Info(ctx, "successfully done weather collecting job", slog.String("timeEstimated", time.Since(start).String()))
	}()

	spanCtx, span := otel.Tracer("").Start(ctx, fmt.Sprintf("[%T.Do]", c))
	defer span.End()

	if err := c.service.CollectData(spanCtx); err != nil {
		logger.Error(ctx, "error in weather collector cron tick", slog.Any("error", err))
		return
	}
}

func (c *Cron) Stop(ctx context.Context) {
	stopCtx := c.cron.Stop()

	select {
	case <-stopCtx.Done():
		logger.Info(ctx, "weather collector cron successfully stopped")
	case <-ctx.Done():
		logger.Warn(ctx, "weather collector cron stop interrupted by context cancellation")
	}
}
