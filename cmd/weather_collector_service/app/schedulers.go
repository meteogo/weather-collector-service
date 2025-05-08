package app

import (
	"context"

	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/internal/closer"
	"github.com/meteogo/weather-collector-service/internal/schedulers/weather_collector_cron"
	"github.com/meteogo/weather-collector-service/internal/schedulers/weather_sender_cron"
	"github.com/robfig/cron/v3"
)

type Schedulers struct {
	WeatherCollectorCron *weather_collector_cron.Cron
}

func InitSchedulers(ctx context.Context, provider config.Provider, services Services) Schedulers {
	c := cron.New()
	weatherCollectorConfig, err := weather_collector_cron.NewConfig(provider)
	if err != nil {
		panic(err)
	}

	weatherCollectorCron := weather_collector_cron.NewCron(weatherCollectorConfig, c, services.WeatherService)
	weatherCollectorCron.Start(ctx)

	weatherSenderConfig, err := weather_sender_cron.NewConfig(provider)
	if err != nil {
		panic(err)
	}

	weatherSenderCron := weather_sender_cron.NewCron(weatherSenderConfig, c, services.WeatherService)
	weatherSenderCron.Start(ctx)

	closer.Add(func(ctx context.Context) error {
		logger.Info(ctx, "stopping weather collector cron")
		weatherCollectorCron.Stop(ctx)
		return nil
	})

	closer.Add(func(ctx context.Context) error {
		logger.Info(ctx, "stopping weather sender cron")
		weatherSenderCron.Stop(ctx)
		return nil
	})

	return Schedulers{
		WeatherCollectorCron: weatherCollectorCron,
	}
}
