package weather_service

import (
	"context"
	"log/slog"
	"sync"

	"github.com/meteogo/logger/pkg/logger"
)

//go:generate mockgen -source service.go -destination service_mocks_test.go -package weather_service_test -typed

type Config interface {
	ReportedCities() ReportedCities
	MonitoringParams() MonitoringParamsMap
	WorkerPoolSize() int
}

type MeteoClient interface {
	CurrentWeather(ctx context.Context, city City, params MonitoringParamsMap) (CityWeatherCondition, error)
}

type Publisher interface {
	PublishConditions(ctx context.Context, conditions CityWeatherConditions) error
}

type Storage interface {
	SaveConditions(ctx context.Context, conditions CityWeatherConditions) error
	GetConditions(ctx context.Context) (CityWeatherConditions, error)
}

type Service struct {
	config      Config
	meteoClient MeteoClient
	publisher   Publisher
	storage     Storage
}

func NewService(config Config, meteoClient MeteoClient, publisher Publisher, storage Storage) *Service {
	return &Service{
		config:      config,
		meteoClient: meteoClient,
		publisher:   publisher,
		storage:     storage,
	}
}

func (s *Service) CollectData(ctx context.Context) error {
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

					weather, err := s.meteoClient.CurrentWeather(ctx, city, s.config.MonitoringParams())
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

	if err := s.storage.SaveConditions(ctx, conditions); err != nil {
		return err
	}

	logger.Info(ctx, "successfully saved reported cities", slog.Int("savedCitiesCount", len(conditions)))
	return nil
}

func (s *Service) SendData(ctx context.Context) error {
	conditions, err := s.storage.GetConditions(ctx)
	if err != nil {
		logger.Error(ctx, "error getting conditions from storage", slog.Any("error", err))
		return err
	}

	if err := s.publisher.PublishConditions(ctx, conditions); err != nil {
		logger.Error(ctx, "error publishing conditions", slog.Any("error", err))
		return err
	}

	logger.Info(ctx, "weather conditions published successfully")
	return nil
}
