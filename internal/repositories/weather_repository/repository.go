package weather_repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/internal/services/weather_service"

	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Save(ctx context.Context, conditions weather_service.CityWeatherConditions) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	qb := psql.
		Insert("current_weather_conditions").
		Columns(
			"city_name",
			"latitude",
			"longitude",
			"captured_at",
			"temperature",
			"relative_humidity_percent",
			"wind_speed",
			"weather_code",
			"cloud_cover_percent",
			"precipitation_millimeters",
			"visibility_meters",
		)

	for _, condition := range conditions {
		qb = qb.Values(
			condition.City.Name,
			condition.City.Coordinates.Lat,
			condition.City.Coordinates.Long,
			condition.CapturedAt,
			condition.Temperature,
			condition.RelativeHumidityPercent,
			condition.WindSpeed,
			condition.WeatherCode,
			condition.CloudCoverPercent,
			condition.PrecipitationMillimeters,
			condition.VisibilityMeters,
		)
	}

	qb = qb.Suffix(`
		ON CONFLICT (city_name)
		DO UPDATE SET
			latitude                  = EXCLUDED.latitude,
    		longitude                 = EXCLUDED.longitude,
    		captured_at               = EXCLUDED.captured_at,
			temperature               = EXCLUDED.temperature,
			relative_humidity_percent = EXCLUDED.relative_humidity_percent,
			wind_speed                = EXCLUDED.wind_speed,
			weather_code              = EXCLUDED.weather_code,
			cloud_cover_percent       = EXCLUDED.cloud_cover_percent,
			precipitation_millimeters = EXCLUDED.precipitation_millimeters,
			visibility_meters         = EXCLUDED.visibility_meters;
	`)

	if _, err := qb.RunWith(r.db).ExecContext(ctx); err != nil {
		logger.Error(ctx, fmt.Sprintf("[%T.Save] unable to ExecContext", r), slog.Any("error", err))
		return err
	}

	return nil
}
