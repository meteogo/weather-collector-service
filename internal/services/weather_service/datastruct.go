package weather_service

import (
	"time"

	"github.com/meteogo/weather-collector-service/internal/pkg/enums"
)

type (
	Coordinates struct {
		Lat  float64
		Long float64
	}

	City struct {
		Name string
		Coordinates
	}

	ReportedCities      []City
	MonitoringParamsMap map[enums.MonitoringParam]string

	CityWeatherCondition struct {
		City                     City
		CapturedAt               time.Time
		Temperature              float64
		RelativeHumidityPercent  uint8
		WindSpeed                float64
		WeatherCode              enums.WeatherCode
		CloudCoverPercent        uint8
		PrecipitationMillimeters enums.Length
		VisibilityMeters         enums.Length
	}

	CityWeatherConditions []CityWeatherCondition
)
