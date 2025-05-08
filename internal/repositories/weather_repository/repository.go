package weather_repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/internal/pkg/enums"
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

func (r *Repository) SaveConditions(ctx context.Context, conditions weather_service.CityWeatherConditions) error {
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
			"visibility_millimeters",
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
			condition.Precipitation,
			condition.Visibility,
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
			visibility_millimeters    = EXCLUDED.visibility_millimeters;
	`)

	if _, err := qb.RunWith(r.db).ExecContext(ctx); err != nil {
		logger.Error(ctx, fmt.Sprintf("[%T.Save] unable to ExecContext", r), slog.Any("error", err))
		return err
	}

	return nil
}

func (r *Repository) GetConditions(ctx context.Context) (weather_service.CityWeatherConditions, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	qb := psql.Select(
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
		"visibility_millimeters",
	).
		From("current_weather_conditions")

	rows, err := qb.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("[%T.GetConditions] QueryContext error", r), slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	var conditions weather_service.CityWeatherConditions

	type intermediate struct {
		CityName                 string
		Latitude                 float64
		Longitude                float64
		CapturedAt               time.Time
		Temperature              float64
		RelativeHumidityPercent  int32
		WindSpeed                float64
		WeatherCode              int64
		CloudCoverPercent        int32
		PrecipitationMillimeters float64
		VisibilityMillimeters    float64
	}

	for rows.Next() {
		var ic intermediate
		if err := rows.Scan(
			&ic.CityName,
			&ic.Latitude,
			&ic.Longitude,
			&ic.CapturedAt,
			&ic.Temperature,
			&ic.RelativeHumidityPercent,
			&ic.WindSpeed,
			&ic.WeatherCode,
			&ic.CloudCoverPercent,
			&ic.PrecipitationMillimeters,
			&ic.VisibilityMillimeters,
		); err != nil {
			logger.Error(ctx, fmt.Sprintf("[%T.GetConditions] Scan error", r), slog.Any("error", err))
			return nil, err
		}

		condition := weather_service.CityWeatherCondition{
			City: weather_service.City{
				Name: ic.CityName,
				Coordinates: weather_service.Coordinates{
					Lat:  ic.Latitude,
					Long: ic.Longitude,
				},
			},
			CapturedAt:              ic.CapturedAt,
			Temperature:             ic.Temperature,
			RelativeHumidityPercent: uint8(ic.RelativeHumidityPercent),
			WindSpeed:               ic.WindSpeed,
			WeatherCode:             enums.WeatherCode(ic.WeatherCode),
			CloudCoverPercent:       uint8(ic.CloudCoverPercent),
			Precipitation:           enums.Length(ic.PrecipitationMillimeters),
			Visibility:              enums.Length(ic.VisibilityMillimeters),
		}

		conditions = append(conditions, condition)
	}

	if err := rows.Err(); err != nil {
		logger.Error(ctx, fmt.Sprintf("[%T.GetConditions] Rows error", r), slog.Any("error", err))
		return nil, err
	}

	return conditions, nil
}
