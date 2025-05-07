package open_meteo_test

import (
	"testing"

	"github.com/meteogo/weather-collector-service/internal/clients/open_meteo"
	"github.com/meteogo/weather-collector-service/internal/pkg/enums"
	"github.com/meteogo/weather-collector-service/internal/services/weather_service"
)

func TestUrlGenerator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		coordinates weather_service.Coordinates
		params      weather_service.MonitoringParamsMap
		expectedURL string
	}{
		{
			name: "happy path",
			coordinates: weather_service.Coordinates{
				Lat:  41.19,
				Long: 4.70,
			},
			params: weather_service.MonitoringParamsMap{
				enums.MonitoringParamTemperature:      "temperature_2m",
				enums.MonitoringParamRelativeHumidity: "relative_humidity_2m",
				enums.MonitoringParamWindSpeed:        "wind_speed_10m",
				enums.MonitoringParamWeatherCode:      "weather_code",
				enums.MonitoringParamCloudCover:       "cloud_cover",
				enums.MonitoringParamPrecipitation:    "precipitation",
				enums.MonitoringParamVisibility:       "visibility",
			},
			expectedURL: "https://api.open-meteo.com/v1/forecast?latitude=41.19&longitude=4.70&current=cloud_cover,precipitation,relative_humidity_2m,temperature_2m,visibility,weather_code,wind_speed_10m",
		},
		{
			name: "one param",
			coordinates: weather_service.Coordinates{
				Lat:  41.19,
				Long: 4.70,
			},
			params: weather_service.MonitoringParamsMap{
				enums.MonitoringParamTemperature: "temperature_2m",
			},
			expectedURL: "https://api.open-meteo.com/v1/forecast?latitude=41.19&longitude=4.70&current=temperature_2m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			generator := open_meteo.NewURLGenerator()
			url := generator.GenerateURL(tt.coordinates, tt.params)
			if url != tt.expectedURL {
				t.Fail()
			}
		})
	}
}
