package weather_publisher

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/internal/pkg/enums"
	"github.com/meteogo/weather-collector-service/internal/services/weather_service"
	weather_collector_events "github.com/meteogo/weather-collector-service/pkg/events"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	weatherMessageKey = "current_weather_conditions"
)

type LibWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type WeatherPublisher struct {
	writer LibWriter
}

func NewPublisher(writer LibWriter) *WeatherPublisher {
	return &WeatherPublisher{
		writer: writer,
	}
}

func (wp *WeatherPublisher) PublishConditions(ctx context.Context, conditions weather_service.CityWeatherConditions) error {
	if len(conditions) == 0 {
		logger.Warn(ctx, "conditions len is zero, skipping publishing")
		return nil
	}

	out := make([]*weather_collector_events.CityWeatherCondition, 0)
	for _, condition := range conditions {
		out = append(out, mapCondition(condition))
	}

	jsonBytes, err := json.Marshal(weather_collector_events.CityWeatherConditions{
		Conditions: out,
	})
	if err != nil {
		logger.Error(ctx, "failed to Marshal conditions to kafka message", slog.Any("error", err))
		return err
	}

	if err := wp.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(weatherMessageKey),
		Value: jsonBytes,
	}); err != nil {
		logger.Error(ctx, "failed to write message to kafka", slog.Any("error", err.Error()))
		return err
	}

	logger.Info(ctx, "weather conditions successfully sent to kafka")
	return nil
}

func mapCondition(c weather_service.CityWeatherCondition) *weather_collector_events.CityWeatherCondition {
	return &weather_collector_events.CityWeatherCondition{
		City: &weather_collector_events.City{
			Name: c.City.Name,
			Coordinates: &weather_collector_events.Coordinates{
				Lat:  c.City.Lat,
				Long: c.City.Long,
			},
		},
		CapturedAt:               timestamppb.New(c.CapturedAt.UTC()),
		Temperature:              c.Temperature,
		RelativeHumidityPercent:  uint32(c.RelativeHumidityPercent),
		WindSpeed:                c.WindSpeed,
		WeatherCode:              weather_collector_events.WeatherCode(c.WeatherCode),
		CloudCoverPercent:        uint32(c.CloudCoverPercent),
		PrecipitationMillimeters: int64(c.Precipitation * enums.Millimeter),
		VisibilityMillimeters:    int64(c.Visibility * enums.Millimeter),
	}
}

func (wp *WeatherPublisher) Close(ctx context.Context) error {
	return wp.writer.Close()
}
