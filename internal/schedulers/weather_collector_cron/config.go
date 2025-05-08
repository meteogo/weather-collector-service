package weather_collector_cron

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/logger/pkg/logger"
	appconfig "github.com/meteogo/weather-collector-service/internal/config"
)

var _ Config = &configImpl{}

type Provider interface {
	config.Provider
}

type ConfigClient interface {
	config.ConfigClient
}

type Value interface {
	config.Value
}

type configImpl struct {
	duration time.Duration

	mu sync.RWMutex
}

func NewConfig(provider Provider) (*configImpl, error) {
	c := &configImpl{
		mu: sync.RWMutex{},
	}

	if err := c.updateDuration(provider.GetConfigClient().GetValue(appconfig.WeatherCollectorCronDuration).Duration()); err != nil {
		logger.Error(context.Background(), "unable to update duration value", slog.Any("error", err))
		return nil, err
	}

	return c, nil
}

func (c *configImpl) updateDuration(duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.duration = duration
	logger.Info(context.Background(), "updated duration value", slog.String(string(appconfig.WeatherCollectorCronDuration), duration.String()))
	return nil
}

func (c *configImpl) Duration() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.duration
}
