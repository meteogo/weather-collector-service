package weather_collector

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/logger/pkg/logger"
	appconfig "github.com/meteogo/weather-collector-service/internal/config"
	"github.com/meteogo/weather-collector-service/internal/pkg/enums"
)

//go:generate mockgen -source config.go -destination config_mocks_test.go -package weather_collector_test -typed

var _ Config = &configImpl{}

type Provider interface {
	config.Provider
}

type ConfigClient interface {
	config.ConfigClient
}

type Value interface {
	config.Value
}

type configImpl struct {
	reportedCities   ReportedCities
	monitoringParams MonitoringParamsMap
	workerPoolSize   int

	mu sync.RWMutex
}

func NewConfig(provider Provider) (*configImpl, error) {
	c := &configImpl{
		reportedCities:   make(ReportedCities, 0),
		monitoringParams: make(MonitoringParamsMap),
		workerPoolSize:   0,

		mu: sync.RWMutex{},
	}

	if err := c.updateReportedCities(provider.GetConfigClient().GetValue(appconfig.ReportedCities).String()); err != nil {
		logger.Error(context.Background(), "unable to update reported cities value", slog.Any("error", err))
		return nil, err
	}

	if err := c.updateMonitoringParams(provider.GetConfigClient().GetValue(appconfig.MonitoringParams).String()); err != nil {
		logger.Error(context.Background(), "unable to update monitoring params value", slog.Any("error", err))
		return nil, err
	}

	if err := c.updateWorkerPoolSize(provider.GetConfigClient().GetValue(appconfig.CollectorWorkerPoolSize).Int()); err != nil {
		logger.Error(context.Background(), "unable to update worker pool size value", slog.Any("error", err))
		return nil, err
	}

	return c, nil
}

func (c *configImpl) updateReportedCities(JSON string) error {
	var cities []struct {
		Name string  `json:"name"`
		Lat  float64 `json:"lat"`
		Long float64 `json:"long"`
	}
	if err := json.Unmarshal([]byte(JSON), &cities); err != nil {
		return err
	}

	reportedCities := make(ReportedCities, 0)
	for _, city := range cities {
		reportedCities = append(reportedCities, City{
			Name: city.Name,
			Coordinates: Coordinates{
				Lat:  city.Lat,
				Long: city.Long,
			},
		})
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.reportedCities = reportedCities
	logger.Info(context.Background(), "updated reported cities value", slog.Any(string(appconfig.ReportedCities), reportedCities))
	return nil
}

func (c *configImpl) updateMonitoringParams(JSON string) error {
	params := make(map[string]string)
	if err := json.Unmarshal([]byte(JSON), &params); err != nil {
		return err
	}

	monitoringParams := make(MonitoringParamsMap)
	for k, v := range params {
		monitoringParams[enums.MonitoringParam(k)] = v
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.monitoringParams = monitoringParams
	logger.Info(context.Background(), "updated monitoring params value", slog.Any(string(appconfig.MonitoringParams), monitoringParams))
	return nil
}

func (c *configImpl) updateWorkerPoolSize(wps int) error {
	if wps < 1 {
		return errors.New("worker pool size value in config can not be less than 1")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.workerPoolSize = wps
	logger.Info(context.Background(), "updated worker pool size value", slog.Int(string(appconfig.CollectorWorkerPoolSize), wps))
	return nil
}

func (c *configImpl) ReportedCities() ReportedCities {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.reportedCities
}

func (c *configImpl) MonitoringParams() MonitoringParamsMap {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.monitoringParams
}

func (c *configImpl) WorkerPoolSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.workerPoolSize
}
