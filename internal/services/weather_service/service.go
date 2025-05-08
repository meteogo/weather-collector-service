package weather_service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/meteogo/logger/pkg/logger"
	"go.opentelemetry.io/otel"
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

type MetricsManager interface {
	AddMeteoClientDurationMetric(ctx context.Context, d time.Duration)
	AddKafkaSendDurationMetric(ctx context.Context, d time.Duration)
}

type Service struct {
	config         Config
	meteoClient    MeteoClient
	publisher      Publisher
	storage        Storage
	metricsManager MetricsManager
}

func NewService(
	config Config,
	meteoClient MeteoClient,
	publisher Publisher,
	storage Storage,
	metricsManager MetricsManager,
) *Service {
	return &Service{
		config:         config,
		meteoClient:    meteoClient,
		publisher:      publisher,
		storage:        storage,
		metricsManager: metricsManager,
	}
}

func (s *Service) CollectData(ctx context.Context) error {
	spanCtx, span := otel.Tracer("").Start(ctx, fmt.Sprintf("[%T.CollectData]", s))
	defer span.End()

	collectStart := time.Now()
	conditions := s.collectDataFromClient(spanCtx)
	s.metricsManager.AddMeteoClientDurationMetric(ctx, time.Since(collectStart))

	if err := s.storage.SaveConditions(spanCtx, conditions); err != nil {
		return err
	}

	logger.Info(ctx, "successfully saved reported cities", slog.Int("savedCitiesCount", len(conditions)))
	return nil
}

func (s *Service) collectDataFromClient(ctx context.Context) CityWeatherConditions {
	_, span := otel.Tracer("").Start(ctx, fmt.Sprintf("[%T.collectDataFromClient]", s))
	defer span.End()

	reportedCities := s.config.ReportedCities()
	if len(reportedCities) == 0 {
		return CityWeatherConditions{}
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

	return conditions
}

func (s *Service) SendData(ctx context.Context) error {
	spanCtx, span := otel.Tracer("").Start(ctx, fmt.Sprintf("[%T.SendData]", s))
	defer span.End()

	conditions, err := s.storage.GetConditions(spanCtx)
	if err != nil {
		logger.Error(ctx, "error getting conditions from storage", slog.Any("error", err))
		return err
	}

	publishStart := time.Now()
	if err := s.publisher.PublishConditions(spanCtx, conditions); err != nil {
		logger.Error(ctx, "error publishing conditions", slog.Any("error", err))
		return err
	}

	s.metricsManager.AddKafkaSendDurationMetric(ctx, time.Since(publishStart))
	logger.Info(ctx, "weather conditions published successfully")
	return nil
}
