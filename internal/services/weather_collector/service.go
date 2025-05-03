package weather_collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/meteogo/logger/pkg/logger"
)

//go:generate mockgen -source service.go -destination service_mocks_test.go -package weather_collector_test -typed

type Config interface {
	ReportedCities() ReportedCities
	MonitoringParams() MonitoringParamsMap
	WorkerPoolSize() int
}

type MeteoClient interface {
	CurrentWeather(ctx context.Context, coordinates Coordinates, params MonitoringParamsMap) (CityWeatherCondition, error)
}

type Storage interface {
	Save(ctx context.Context, conditions CityWeatherConditions) error
}

type Service struct {
	config      Config
	meteoClient MeteoClient
	storage     Storage
}

func NewService(config Config, meteoClient MeteoClient, storage Storage) *Service {
	return &Service{
		config:      config,
		meteoClient: meteoClient,
		storage:     storage,
	}
}

func (s *Service) CollectData(ctx context.Context) error {
	start := time.Now()
	reportedCities := s.config.ReportedCities()
	if len(reportedCities) == 0 {
		return nil
	}

	cityChan := make(chan City, len(reportedCities))
	resultChan := make(chan CityWeatherCondition, len(reportedCities))

	var (
		wg         sync.WaitGroup
		resultWg   sync.WaitGroup
		mu         sync.Mutex
		conditions CityWeatherConditions
	)

	for _, city := range reportedCities {
		cityChan <- city
	}
	close(cityChan)

	for i := 0; i < s.config.WorkerPoolSize() && i < len(reportedCities); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case city, ok := <-cityChan:
					if !ok {
						return
					}

					weather, err := s.meteoClient.CurrentWeather(ctx, city.Coordinates, s.config.MonitoringParams())
					if err != nil {
						logger.Error(ctx, "unable to get current weather for city", slog.Any("city", city), slog.Any("err", err))
						continue
					}

					resultChan <- weather
				}
			}
		}()
	}

	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for result := range resultChan {
			mu.Lock()
			conditions = append(conditions, result)
			mu.Unlock()
		}
	}()

	wg.Wait()
	close(resultChan)
	resultWg.Wait()

	if err := s.storage.Save(ctx, conditions); err != nil {
		return err
	}

	logger.Info(
		ctx,
		"successfully done weather collecting job",
		slog.Int("citiesCount", len(s.config.ReportedCities())),
		slog.Any("timeEstimated", time.Since(start)),
	)
	return nil
}
