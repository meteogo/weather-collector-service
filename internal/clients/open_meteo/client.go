package open_meteo

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/internal/pkg/enums"
	"github.com/meteogo/weather-collector-service/internal/services/weather_service"
)

type OpenMeteoURLGenerator interface {
	GenerateURL(coordinates weather_service.Coordinates, params weather_service.MonitoringParamsMap) string
}

type Client struct {
	urlGenerator OpenMeteoURLGenerator
}

func NewOpenMeteoClient(urlGenerator OpenMeteoURLGenerator) *Client {
	return &Client{
		urlGenerator: urlGenerator,
	}
}

func (c *Client) CurrentWeather(ctx context.Context, city weather_service.City, params weather_service.MonitoringParamsMap) (weather_service.CityWeatherCondition, error) {
	url := c.urlGenerator.GenerateURL(city.Coordinates, params)

	resp, err := http.Get(url)
	if err != nil {
		logger.Error(ctx, "unable to http.Get", slog.Any("coords", city.Coordinates), slog.Any("params", params), slog.Any("error", err))
		return weather_service.CityWeatherCondition{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(ctx, "unable to read response body", slog.Any("coords", city.Coordinates), slog.Any("error", err))
		return weather_service.CityWeatherCondition{}, err
	}

	type CurrentWeatherResponse struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Current   struct {
			Time               string  `json:"time"`
			Temperature2m      float64 `json:"temperature_2m"`
			RelativeHumidity2m uint8   `json:"relative_humidity_2m"`
			WindSpeed10m       float64 `json:"wind_speed_10m"`
			WeatherCode        int     `json:"weather_code"`
			CloudCover         uint8   `json:"cloud_cover"`
			Precipitation      float64 `json:"precipitation"`
			Visibility         float64 `json:"visibility"`
		} `json:"current"`
	}

	var response CurrentWeatherResponse
	if err := json.Unmarshal(body, &response); err != nil {
		logger.Error(ctx, "unable to unmarshal response", slog.Any("coords", city.Coordinates), slog.Any("error", err))
		return weather_service.CityWeatherCondition{}, err
	}

	capturedAt, err := time.Parse("2006-01-02T15:04", response.Current.Time)
	if err != nil {
		logger.Error(ctx, "unable to parse time", slog.Any("time", response.Current.Time), slog.Any("error", err))
		return weather_service.CityWeatherCondition{}, err
	}

	return weather_service.CityWeatherCondition{
		City: weather_service.City{
			Name: city.Name,
			Coordinates: weather_service.Coordinates{
				Lat:  response.Latitude,
				Long: response.Longitude,
			},
		},
		CapturedAt:              capturedAt,
		Temperature:             response.Current.Temperature2m,
		RelativeHumidityPercent: response.Current.RelativeHumidity2m,
		WindSpeed:               response.Current.WindSpeed10m,
		WeatherCode:             enums.WeatherCode(response.Current.WeatherCode),
		CloudCoverPercent:       response.Current.CloudCover,
		Precipitation:           enums.Length(response.Current.Precipitation) * enums.Millimeter,
		Visibility:              enums.Length(response.Current.Visibility) * enums.Meter,
	}, nil
}
