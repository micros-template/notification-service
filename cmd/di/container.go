package di

import (
	"10.1.20.130/dropping/notification-service/config/logger"
	mail "10.1.20.130/dropping/notification-service/config/mail"
	mq "10.1.20.130/dropping/notification-service/config/message-queue"
	"10.1.20.130/dropping/notification-service/internal/domain/handler"
	"10.1.20.130/dropping/notification-service/internal/domain/service"
	_mail "10.1.20.130/dropping/notification-service/internal/infrastructure/mail"
	_mq "10.1.20.130/dropping/notification-service/internal/infrastructure/message-queue"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/dig"
)

func BuildContainer() *dig.Container {
	container := dig.New()

	// logger
	if err := container.Provide(logger.New); err != nil {
		panic("Failed to provide logger: " + err.Error())
	}
	// nats connection
	if err := container.Provide(mq.New); err != nil {
		panic("Failed to provide message queue: " + err.Error())
	}
	// mail dialer
	if err := container.Provide(mail.New); err != nil {
		panic("Failed to provide mail dialer: " + err.Error())
	}
	// mail infra
	if err := container.Provide(_mail.New); err != nil {
		panic("Failed to provide mail dialer: " + err.Error())
	}
	// jetstream
	if err := container.Provide(jetstream.New); err != nil {
		panic("Failed to provide jetstream: " + err.Error())
	}
	// nats infra
	if err := container.Provide(_mq.NewNatsInfrastructure); err != nil {
		panic("Failed to provide message queue infra: " + err.Error())
	}
	// subscriber service
	if err := container.Provide(service.NewSubscriberService); err != nil {
		panic("Failed to provide subscriber service " + err.Error())
	}
	// subscriber handler
	if err := container.Provide(handler.NewSubscriberHandler); err != nil {
		panic("Failed to provide subscriber handler" + err.Error())
	}
	return container
}
