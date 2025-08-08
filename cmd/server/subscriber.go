package server

import (
	"context"
	"log"

	"10.1.20.130/dropping/log-management/pkg"
	ld "10.1.20.130/dropping/log-management/pkg/dto"
	"10.1.20.130/dropping/notification-service/internal/domain/handler"
	mq "10.1.20.130/dropping/notification-service/internal/infrastructure/message-queue"
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
	) {
		defer _mq.Drain()
		err := mq.CreateOrUpdateNewStream(ctx, &jetstream.StreamConfig{
			Name:        viper.GetString("jetstream.notification.stream.name"),
			Description: viper.GetString("jetstream.notification.stream.description"),
			Subjects:    []string{viper.GetString("jetstream.notification.subject.global")},
			MaxBytes:    10 * 1024 * 1024,
			Storage:     jetstream.FileStorage,
		})
		if err != nil {
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
			logger.Fatal().Err(err).Msg("Failed to create or update mail consumer")
		}

		_, err = emailCons.Consume(func(msg jetstream.Msg) {
			logEmitter.EmitLog(context.Background(), ld.LogMessage{
				Type:     "access",
				Service:  "notification_service",
				Msg:      string(msg.Data()),
				Protocol: "PUB-SUB",
			})
			go func() {
				sh.EmailHandler(msg)
				msg.Ack()
			}()
		})

		if err != nil {
			logger.Error().Err(err).Msg("Failed to consume email consumer")
			return
		}

		if s.ConnectionReady != nil {
			s.ConnectionReady <- true
		}

		logger.Info().Msg("subscriber for notification is running")
		<-ctx.Done()
		logger.Info().Msg("Shutting down subscriber for notification")
	})
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
}
