package app

import (
	"context"
	"log/slog"
	"net"

	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/internal/closer"
	appconfig "github.com/meteogo/weather-collector-service/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracer(ctx context.Context, provider config.Provider) {
	var (
		host = provider.GetSecretClient().GetSecret(appconfig.JaegerHost).String()
		port = provider.GetSecretClient().GetSecret(appconfig.JaegerPort).String()
	)

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(net.JoinHostPort(host, port)),
	)
	if err != nil {
		logger.Error(ctx, "failed to create new tracer", slog.Any("error", err.Error()))
		panic(err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(provider.GetConfigClient().GetValue(appconfig.ApplicationName).String()),
			semconv.DeploymentEnvironmentKey.String(provider.GetConfigClient().GetValue(appconfig.Env).String()),
		),
	)
	if err != nil {
		logger.Error(ctx, "error while creating a tracer resource", slog.Any("error", err.Error()))
		panic(err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	logger.Info(ctx, "tracer provider created successfully")
	closer.Add(func(ctx context.Context) error {
		logger.Info(ctx, "shutting down tracer provider")
		return tp.Shutdown(ctx)
	})
}
