package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/internal/closer"
	appconfig "github.com/meteogo/weather-collector-service/internal/config"
	"github.com/meteogo/weather-collector-service/internal/repositories/weather_repository"
)

type Repositories struct {
	WeatherRepo *weather_repository.Repository
}

func InitRepositories(ctx context.Context, provider config.Provider) Repositories {
	var (
		user         = provider.GetSecretClient().GetSecret(appconfig.PostgresUser).String()
		password     = provider.GetSecretClient().GetSecret(appconfig.PostgresPassword).String()
		host         = provider.GetSecretClient().GetSecret(appconfig.PostgresHost).String()
		port         = provider.GetSecretClient().GetSecret(appconfig.PostgresPort).String()
		databaseName = provider.GetSecretClient().GetSecret(appconfig.PostgresDatabaseName).String()
	)

	connection := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=disable host=%s port=%s",
		user,
		password,
		databaseName,
		host,
		port,
	)

	db, err := sql.Open("postgres", connection)
	if err != nil {
		logger.Error(ctx, "failed to open postgres connection", slog.Any("error", err))
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		logger.Error(ctx, "failed to ping postgres database", slog.Any("error", err))
		panic(err)
	}

	closer.Add(func(ctx context.Context) error {
		logger.Info(ctx, "closing database connection")
		return db.Close()
	})

	logger.Info(ctx, "database connection established successfully")
	return Repositories{
		WeatherRepo: weather_repository.NewRepository(db),
	}
}
