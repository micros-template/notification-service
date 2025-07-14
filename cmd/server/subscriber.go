package server

import (
	"context"
	"log"

	"github.com/dropboks/notification-service/internal/domain/handler"
	mq "github.com/dropboks/notification-service/internal/infrastructure/message-queue"
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
	) {
		defer _mq.Drain()
		err := mq.CreateOrUpdateNewStream(ctx, &jetstream.StreamConfig{
			Name:        viper.GetString("jetstream.stream.name"),
			Description: viper.GetString("jetstream.stream.description"),
			Subjects:    []string{viper.GetString("jetstream.subject.global")},
			MaxBytes:    10 * 1024 * 1024,
			Storage:     jetstream.FileStorage,
		})
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create or update notification stream")
		}

		// consumer for email
		emailCons, err := mq.CreateOrUpdateNewConsumer(ctx, viper.GetString("jetstream.stream.name"), &jetstream.ConsumerConfig{
			Name:          viper.GetString("jetstream.consumer.mail"),
			Durable:       viper.GetString("jetstream.consumer.mail"),
			FilterSubject: viper.GetString("jetstream.subject.mail"),
			AckPolicy:     jetstream.AckExplicitPolicy,
			DeliverPolicy: jetstream.DeliverNewPolicy,
		})

		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create or update mail consumer")
		}

		_, err = emailCons.Consume(func(msg jetstream.Msg) {
			logger.Info().
				Str("subject", msg.Subject()).
				Msgf("Received message: %s", string(msg.Data()))
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
