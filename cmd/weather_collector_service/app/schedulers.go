package app

import (
	"context"

	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/weather-collector-service/internal/schedulers/weather_cron"
	"github.com/robfig/cron/v3"
)

type Schedulers struct {
	WeatherCollectorCron *weather_cron.Cron
}

func InitSchedulers(ctx context.Context, provider config.Provider, services Services) Schedulers {
	c := cron.New()
	weatherCollectorConfig, err := weather_cron.NewConfig(provider)
	if err != nil {
		panic(err)
	}

	weatherCollectorCron := weather_cron.NewCron(weatherCollectorConfig, c, services.WeatherService)
	weatherCollectorCron.Start(ctx)
	return Schedulers{
		WeatherCollectorCron: weatherCollectorCron,
	}
}
