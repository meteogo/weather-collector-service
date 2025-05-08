package weather_service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/internal/pkg/enums"
	"github.com/meteogo/weather-collector-service/internal/services/weather_service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	berlinCondition = weather_service.CityWeatherCondition{
		City: weather_service.City{
			Name: "Berlin",
			Coordinates: weather_service.Coordinates{
				Lat:  52.52,
				Long: 13.41,
			},
		},
		CapturedAt:              must(time.ParseInLocation(time.DateTime, "2025-05-03 12:58:00", time.FixedZone("", 0))),
		Temperature:             12.5,
		RelativeHumidityPercent: 20,
		WindSpeed:               3.7,
		WeatherCode:             enums.ClearSky,
		CloudCoverPercent:       3,
		Precipitation:           3 * enums.Millimeter,
		Visibility:              40 * enums.Kilometer,
	}

	parisCondition = weather_service.CityWeatherCondition{
		City: weather_service.City{
			Name: "Paris",
			Coordinates: weather_service.Coordinates{
				Lat:  48.86,
				Long: 2.35,
			},
		},
		CapturedAt:              must(time.ParseInLocation(time.DateTime, "2025-05-03 13:03:00", time.FixedZone("", 0))),
		Temperature:             10.5,
		RelativeHumidityPercent: 13,
		WindSpeed:               4.2,
		WeatherCode:             enums.PartlyCloudy,
		CloudCoverPercent:       29,
		Precipitation:           15 * enums.Millimeter,
		Visibility:              31 * enums.Kilometer,
	}

	londonCondition = weather_service.CityWeatherCondition{
		City: weather_service.City{
			Name: "London",
			Coordinates: weather_service.Coordinates{
				Lat:  41.90,
				Long: -0.13,
			},
		},
		CapturedAt:              must(time.ParseInLocation(time.DateTime, "2025-05-03 13:05:00", time.FixedZone("", 0))),
		Temperature:             11.8,
		RelativeHumidityPercent: 20,
		WindSpeed:               6.2,
		WeatherCode:             enums.Fog,
		CloudCoverPercent:       75,
		Precipitation:           40 * enums.Millimeter,
		Visibility:              5 * enums.Kilometer,
	}
)

func TestWeatherService_CollectData(t *testing.T) {
	t.Parallel()
	logger.InitLogger(logger.EnvTypeTesting, slog.LevelDebug)

	tests := []struct {
		name        string
		config      func(ctrl *gomock.Controller) weather_service.Config
		meteoClient func(ctrl *gomock.Controller) weather_service.MeteoClient
		storage     func(ctrl *gomock.Controller) weather_service.Storage
		wantErrFunc assert.ErrorAssertionFunc
	}{
		{
			name: "happy path",
			config: func(ctrl *gomock.Controller) weather_service.Config {
				return mockConfig(ctrl)
			},
			meteoClient: func(ctrl *gomock.Controller) weather_service.MeteoClient {
				mock := NewMockMeteoClient(ctrl)
				mock.EXPECT().
					CurrentWeather(gomock.Any(), gomock.Eq(weather_service.City{
						Name: "Berlin",
						Coordinates: weather_service.Coordinates{
							Lat:  52.52,
							Long: 13.41,
						},
					}), gomock.Any()).
					Return(berlinCondition, nil).
					Times(1)

				mock.EXPECT().
					CurrentWeather(gomock.Any(), gomock.Eq(weather_service.City{
						Name: "Paris",
						Coordinates: weather_service.Coordinates{
							Lat:  48.86,
							Long: 2.35,
						},
					}), gomock.Any()).
					Return(parisCondition, nil).
					Times(1)

				mock.EXPECT().
					CurrentWeather(gomock.Any(), gomock.Eq(weather_service.City{
						Name: "London",
						Coordinates: weather_service.Coordinates{
							Lat:  41.90,
							Long: -0.13,
						},
					}), gomock.Any()).
					Return(londonCondition, nil).
					Times(1)

				return mock
			},
			storage: func(ctrl *gomock.Controller) weather_service.Storage {
				mock := NewMockStorage(ctrl)
				expectedConditions := weather_service.CityWeatherConditions{
					berlinCondition,
					parisCondition,
					londonCondition,
				}

				mock.EXPECT().
					SaveConditions(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, conditions weather_service.CityWeatherConditions) error {
						assert.ElementsMatch(t, expectedConditions, conditions, "Saved conditions do not match expected conditions")
						return nil
					}).
					Times(1)

				return mock
			},
			wantErrFunc: assert.NoError,
		},
		{
			name: "meteo client error",
			config: func(ctrl *gomock.Controller) weather_service.Config {
				return mockConfig(ctrl)
			},
			meteoClient: func(ctrl *gomock.Controller) weather_service.MeteoClient {
				mock := NewMockMeteoClient(ctrl)
				mock.EXPECT().
					CurrentWeather(gomock.Any(), gomock.Eq(weather_service.City{
						Name: "Berlin",
						Coordinates: weather_service.Coordinates{
							Lat:  52.52,
							Long: 13.41,
						},
					}), gomock.Any()).
					Return(berlinCondition, nil).
					Times(1)

				mock.EXPECT().
					CurrentWeather(gomock.Any(), gomock.Eq(weather_service.City{
						Name: "Paris",
						Coordinates: weather_service.Coordinates{
							Lat:  48.86,
							Long: 2.35,
						},
					}), gomock.Any()).
					Return(parisCondition, nil).
					Times(1)

				mock.EXPECT().
					CurrentWeather(gomock.Any(), gomock.Eq(weather_service.City{
						Name: "London",
						Coordinates: weather_service.Coordinates{
							Lat:  41.90,
							Long: -0.13,
						},
					}), gomock.Any()).
					Return(weather_service.CityWeatherCondition{}, errors.New("client error")).
					Times(1)

				return mock
			},
			storage: func(ctrl *gomock.Controller) weather_service.Storage {
				mock := NewMockStorage(ctrl)
				expectedConditions := weather_service.CityWeatherConditions{
					berlinCondition,
					parisCondition,
				}

				mock.EXPECT().
					SaveConditions(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, conditions weather_service.CityWeatherConditions) error {
						assert.ElementsMatch(t, expectedConditions, conditions, "Saved conditions do not match expected conditions")
						return nil
					}).
					Times(1)

				return mock
			},
			wantErrFunc: assert.NoError,
		},
		{
			name: "storage error",
			config: func(ctrl *gomock.Controller) weather_service.Config {
				return mockConfig(ctrl)
			},
			meteoClient: func(ctrl *gomock.Controller) weather_service.MeteoClient {
				mock := NewMockMeteoClient(ctrl)
				mock.EXPECT().
					CurrentWeather(gomock.Any(), gomock.Eq(weather_service.City{
						Name: "Berlin",
						Coordinates: weather_service.Coordinates{
							Lat:  52.52,
							Long: 13.41,
						},
					}), gomock.Any()).
					Return(berlinCondition, nil).
					Times(1)

				mock.EXPECT().
					CurrentWeather(gomock.Any(), gomock.Eq(weather_service.City{
						Name: "Paris",
						Coordinates: weather_service.Coordinates{
							Lat:  48.86,
							Long: 2.35,
						},
					}), gomock.Any()).
					Return(parisCondition, nil).
					Times(1)

				mock.EXPECT().
					CurrentWeather(gomock.Any(), gomock.Eq(weather_service.City{
						Name: "London",
						Coordinates: weather_service.Coordinates{
							Lat:  41.90,
							Long: -0.13,
						},
					}), gomock.Any()).
					Return(londonCondition, nil).
					Times(1)

				return mock
			},
			storage: func(ctrl *gomock.Controller) weather_service.Storage {
				mock := NewMockStorage(ctrl)
				expectedConditions := weather_service.CityWeatherConditions{
					berlinCondition,
					parisCondition,
					londonCondition,
				}

				mock.EXPECT().
					SaveConditions(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, conditions weather_service.CityWeatherConditions) error {
						assert.ElementsMatch(t, expectedConditions, conditions, "Saved conditions do not match expected conditions")
						return errors.New("storage error")
					}).
					Times(1)

				return mock
			},
			wantErrFunc: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			service := weather_service.NewService(
				tt.config(ctrl),
				tt.meteoClient(ctrl),
				nil,
				tt.storage(ctrl),
			)

			err := service.CollectData(context.Background())
			if !tt.wantErrFunc(t, err) {
				t.Fail()
			}
		})
	}
}

func mockConfig(ctrl *gomock.Controller) weather_service.Config {
	mock := NewMockConfig(ctrl)
	mock.EXPECT().
		ReportedCities().
		Return(weather_service.ReportedCities{
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
					Lat:  41.90,
					Long: -0.13,
				},
			},
		}).
		AnyTimes()

	mock.EXPECT().
		MonitoringParams().
		Return(weather_service.MonitoringParamsMap{
			enums.MonitoringParamTemperature:      "temperature_2m",
			enums.MonitoringParamRelativeHumidity: "relative_humidity_2m",
			enums.MonitoringParamWindSpeed:        "wind_speed_10m",
			enums.MonitoringParamWeatherCode:      "weather_code",
			enums.MonitoringParamCloudCover:       "cloud_cover",
			enums.MonitoringParamPrecipitation:    "precipitation",
			enums.MonitoringParamVisibility:       "visibility",
		}).
		AnyTimes()

	mock.EXPECT().
		WorkerPoolSize().
		Return(3).
		AnyTimes()
	return mock
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
