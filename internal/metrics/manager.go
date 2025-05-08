package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/meteogo/logger/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Manager struct {
	openMeteoFetchDuration prometheus.Histogram
	kafkaSendDuration      prometheus.Histogram
}

func NewManager() *Manager {
	return &Manager{
		openMeteoFetchDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "open_meteo_featch_duration_ms",
		}),

		kafkaSendDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "kafka_send_duration_ms",
		}),
	}
}

func (m *Manager) Start(ctx context.Context) {
	logger.Info(ctx, fmt.Sprintf("[%T.Start] starting a metrics manager", m))

	prometheus.MustRegister(m.openMeteoFetchDuration)
	prometheus.MustRegister(m.kafkaSendDuration)

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)
}

func (m *Manager) AddMeteoClientDurationMetric(ctx context.Context, d time.Duration) {
	m.openMeteoFetchDuration.Observe(float64(d.Milliseconds()))
}
func (m *Manager) AddKafkaSendDurationMetric(ctx context.Context, d time.Duration) {
	m.kafkaSendDuration.Observe(float64(d.Milliseconds()))
}
