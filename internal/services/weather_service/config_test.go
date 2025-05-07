package weather_service_test

import (
	"testing"

	"github.com/meteogo/config/pkg/config"
	appconfig "github.com/meteogo/weather-collector-service/internal/config"
	"github.com/meteogo/weather-collector-service/internal/pkg/enums"
	"github.com/meteogo/weather-collector-service/internal/services/weather_service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWeatherCollectorConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		wantReportedCities   weather_service.ReportedCities
		wantMonitoringParams weather_service.MonitoringParamsMap
		wantWorkerPoolSize   int
		provider             func(ctrl *gomock.Controller) config.Provider
		wantErrFunc          assert.ErrorAssertionFunc
	}{
		{
			name: "happy path",
			wantReportedCities: weather_service.ReportedCities{
				{
					Name: "Berlin",
					Coordinates: weather_service.Coordinates{
						Lat:  52.52,
						Long: 13.41,
					},
				},
				{
					Name: "Paris",
					Coordinates: weather_service.Coordinates{
						Lat:  48.86,
						Long: 2.35,
					},
				},
				{
					Name: "London",
					Coordinates: weather_service.Coordinates{
						Lat:  51.51,
						Long: -0.13,
					},
				},
			},
			wantMonitoringParams: weather_service.MonitoringParamsMap{
				enums.MonitoringParamTemperature:      "temperature_2m",
				enums.MonitoringParamRelativeHumidity: "relative_humidity_2m",
				enums.MonitoringParamWindSpeed:        "wind_speed_10m",
				enums.MonitoringParamWeatherCode:      "weather_code",
				enums.MonitoringParamCloudCover:       "cloud_cover",
				enums.MonitoringParamPrecipitation:    "precipitation",
				enums.MonitoringParamVisibility:       "visibility",
			},
			wantWorkerPoolSize: 10,
			provider: func(ctrl *gomock.Controller) config.Provider {
				return mockProvider(ctrl)
			},
			wantErrFunc: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			cfg, err := weather_service.NewConfig(tt.provider(ctrl))
			if !tt.wantErrFunc(t, err) {
				t.Fail()
			}

			assert.Equal(t, tt.wantReportedCities, cfg.ReportedCities())
			assert.Equal(t, tt.wantMonitoringParams, cfg.MonitoringParams())
			assert.Equal(t, tt.wantWorkerPoolSize, cfg.WorkerPoolSize())
		})
	}
}

func mockProvider(crtl *gomock.Controller) config.Provider {
	providerMock := NewMockProvider(crtl)
	clientMock := NewMockConfigClient(crtl)

	providerMock.EXPECT().
		GetConfigClient().
		Return(clientMock).
		AnyTimes()

	{
		reportedCitiesValueMock := NewMockValue(crtl)
		reportedCitiesValueMock.EXPECT().
			String().
			Return(`
			[
				{
					"name": "Berlin",
					"lat": 52.52,
					"long": 13.41
				},
				{
					"name": "Paris",
					"lat": 48.86,
					"long": 2.35
				},
				{
					"name": "London",
					"lat": 51.51,
					"long": -0.13
				}
			]
			`).
			Times(1)

		clientMock.EXPECT().
			GetValue(gomock.Eq(appconfig.ReportedCities)).
			Return(reportedCitiesValueMock).
			Times(1)
	}

	{
		monitoringParamsValueMock := NewMockValue(crtl)
		monitoringParamsValueMock.EXPECT().
			String().
			Return(`
			{
				"temperature": "temperature_2m",
				"relativeHumidity": "relative_humidity_2m",
				"windSpeed": "wind_speed_10m",
				"weatherCode": "weather_code",
				"cloudCover": "cloud_cover",
				"precipitation": "precipitation",
				"visibility": "visibility"
			}
			`).
			Times(1)

		clientMock.EXPECT().
			GetValue(gomock.Eq(appconfig.MonitoringParams)).
			Return(monitoringParamsValueMock).
			Times(1)
	}

	{
		workerPoolValueMock := NewMockValue(crtl)
		workerPoolValueMock.EXPECT().
			Int().
			Return(10).
			Times(1)

		clientMock.EXPECT().
			GetValue(gomock.Eq(appconfig.CollectorWorkerPoolSize)).
			Return(workerPoolValueMock).
			Times(1)
	}

	return providerMock
}
