package server

import (
	"context"
	"fmt"
	"log"

	"github.com/micros-template/log-service/pkg"
	ld "github.com/micros-template/log-service/pkg/dto"
	"github.com/micros-template/notification-service/internal/domain/handler"
	"github.com/micros-template/notification-service/internal/infrastructure/logger"
	mq "github.com/micros-template/notification-service/internal/infrastructure/message-queue"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"go.uber.org/dig"
)

type Subscriber struct {
	Container       *dig.Container
	ConnectionReady chan bool
}

func (s *Subscriber) Run(ctx context.Context) {
	err := s.Container.Invoke(func(
		logger zerolog.Logger,
		sh handler.SubscriberHandler,
		js jetstream.JetStream,
		mq mq.Nats,
		_mq *nats.Conn,
		logEmitter pkg.LogEmitter,
		logEmitterInfra logger.LoggerInfra,
	) {
		defer func() {
			if err := _mq.Drain(); err != nil {
				logger.Error().Err(err).Msg("Failed to drain nats client")
			}
		}()
		err := mq.CreateOrUpdateNewStream(ctx, &jetstream.StreamConfig{
			Name:        viper.GetString("jetstream.notification.stream.name"),
			Description: viper.GetString("jetstream.notification.stream.description"),
			Subjects:    []string{viper.GetString("jetstream.notification.subject.global")},
			MaxBytes:    10 * 1024 * 1024,
			Storage:     jetstream.FileStorage,
		})
		if err != nil {
			go func() {
				if err := logEmitterInfra.EmitLog("ERR", fmt.Sprintf("Failed to create or update notification stream: %v", err.Error())); err != nil {
					logger.Error().Err(err).Msg("failed to emit log")
				}
			}()
			logger.Fatal().Err(err).Msg("Failed to create or update notification stream")
		}

		// consumer for email
		emailCons, err := mq.CreateOrUpdateNewConsumer(ctx, viper.GetString("jetstream.notification.stream.name"), &jetstream.ConsumerConfig{
			Name:          viper.GetString("jetstream.notification.consumer.mail"),
			Durable:       viper.GetString("jetstream.notification.consumer.mail"),
			FilterSubject: viper.GetString("jetstream.notification.subject.mail"),
			AckPolicy:     jetstream.AckExplicitPolicy,
			DeliverPolicy: jetstream.DeliverNewPolicy,
		})

		if err != nil {
			go func() {
				if err := logEmitterInfra.EmitLog("ERR", fmt.Sprintf("Failed to create or update mail consumer: %v", err.Error())); err != nil {
					logger.Error().Err(err).Msg("failed to emit log")
				}
			}()
			logger.Fatal().Err(err).Msg("Failed to create or update mail consumer")
		}

		_, err = emailCons.Consume(func(msg jetstream.Msg) {
			go func() {
				if err := logEmitter.EmitLog(context.Background(), ld.LogMessage{
					Type:     "INFO",
					Service:  "notification_service",
					Msg:      string(msg.Data()),
					Protocol: "PUB-SUB",
				}); err != nil {
					logger.Error().Err(err).Msg("failed to emit log")
				}
			}()
			go func() {
				_ = sh.EmailHandler(msg)
				if err := msg.Ack(); err != nil {
					go func() {
						if err := logEmitterInfra.EmitLog("ERR", fmt.Sprintf("Error acknowledging message: %v", err.Error())); err != nil {
							logger.Error().Err(err).Msg("failed to emit log")
						}
					}()
					logger.Error().Err(err).Msg("Error acknowledging message")
				}
			}()
		})

		if err != nil {
			go func() {
				if err := logEmitterInfra.EmitLog("ERR", fmt.Sprintf("Failed to consume email consumer: %v", err.Error())); err != nil {
					logger.Error().Err(err).Msg("failed to emit log")
				}
			}()
			logger.Error().Err(err).Msg("Failed to consume email consumer")
			return
		}

		if s.ConnectionReady != nil {
			s.ConnectionReady <- true
		}
		go func() {
			if err := logEmitterInfra.EmitLog("INFO", "subscriber for notification is running"); err != nil {
				logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		logger.Info().Msg("subscriber for notification is running")

		<-ctx.Done()
		go func() {
			if err := logEmitterInfra.EmitLog("INFO", "Shutting down subscriber for notification"); err != nil {
				logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		logger.Info().Msg("Shutting down subscriber for notification")
	})
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
}
