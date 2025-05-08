package app

import (
	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/weather-collector-service/internal/services/weather_service"
)

type Services struct {
	WeatherService *weather_service.Service
}

func InitServices(provider config.Provider, clients Clients, publishers Publishers, repositories Repositories) Services {
	weatherServiceConfig, err := weather_service.NewConfig(provider)
	if err != nil {
		panic(err)
	}

	return Services{
		WeatherService: weather_service.NewService(weatherServiceConfig, clients.openMeteoClient, publishers.weather, repositories.WeatherRepo),
	}
}
