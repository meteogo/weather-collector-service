package app

import (
	"context"
	"net"

	"github.com/meteogo/config/pkg/config"
	"github.com/meteogo/logger/pkg/logger"
	"github.com/meteogo/weather-collector-service/internal/closer"
	appconfig "github.com/meteogo/weather-collector-service/internal/config"
	weather_publisher "github.com/meteogo/weather-collector-service/internal/kafka/publisher/weather"
	"github.com/segmentio/kafka-go"
)

type Publishers struct {
	weather *weather_publisher.WeatherPublisher
}

func InitPublishers(ctx context.Context, provider config.Provider) Publishers {
	var (
		host          = provider.GetSecretClient().GetSecret(appconfig.KafkaHost).String()
		port          = provider.GetSecretClient().GetSecret(appconfig.KafkaPort).String()
		weatherTopic  = provider.GetSecretClient().GetSecret(appconfig.KafkaWeatherTopicName).String()
		weatherWriter = &kafka.Writer{
			Addr:     kafka.TCP(net.JoinHostPort(host, port)),
			Topic:    weatherTopic,
			Balancer: &kafka.LeastBytes{},
		}
		weatherPublisher = weather_publisher.NewPublisher(weatherWriter)
	)

	mustConnectToKafka(ctx, host, port)
	closer.Add(func(ctx context.Context) error {
		logger.Info(ctx, "closing weather publisher")
		return weatherPublisher.Close(ctx)
	})
	return Publishers{
		weather: weatherPublisher,
	}
}

func mustConnectToKafka(ctx context.Context, host, port string) {
	conn, err := kafka.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		logger.Error(ctx, "can not establish connection with kafka")
		panic(err)
	}

	logger.Info(ctx, "connection to kafka established successfully")
	defer conn.Close()
}
