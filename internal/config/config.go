package appconfig

import "github.com/meteogo/config/pkg/config"

const (
	WeatherCollectorCronDuration = config.Key("weather_collector_cron_duration")
	CollectorWorkerPoolSize      = config.Key("collector_worker_pool_size")
	ReportedCities               = config.Key("reported_cities")
	MonitoringParams             = config.Key("monitoring_params")
)

const (
	PostgresUser         = config.Secret("POSTGRES_USER")
	PostgresPassword     = config.Secret("POSTGRES_PASSWORD")
	PostgresHost         = config.Secret("POSTGRES_HOST")
	PostgresPort         = config.Secret("POSTGRES_PORT")
	PostgresDatabaseName = config.Secret("POSTGRES_DATABASE_NAME")
)
