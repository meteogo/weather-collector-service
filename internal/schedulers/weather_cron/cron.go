package weather_cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/meteogo/logger/pkg/logger"
	"github.com/robfig/cron/v3"
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
	if err := c.service.CollectData(ctx); err != nil {
		logger.Error(ctx, "error in weather collector cron tick", slog.Any("error", err))
		return
	}

	logger.Info(ctx, "successfully done weather collecting job", slog.String("timeEstimated", time.Since(start).String()))
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
