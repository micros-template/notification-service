package handler

import (
	"encoding/json"

	"10.1.20.130/dropping/notification-service/internal/domain/service"
	"github.com/dropboks/sharedlib/dto"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
)

type (
	SubscriberHandler interface {
		EmailHandler(msg jetstream.Msg)
	}
	subscriberHandler struct {
		subsService service.SubscriberService
		logger      zerolog.Logger
	}
)

func NewSubscriberHandler(svc service.SubscriberService, logger zerolog.Logger) SubscriberHandler {
	return &subscriberHandler{
		subsService: svc,
		logger:      logger,
	}
}

func (s *subscriberHandler) EmailHandler(msg jetstream.Msg) {
	var msgData dto.MailNotificationMessage
	err := json.Unmarshal(msg.Data(), &msgData)
	if err != nil {
		s.logger.Error().Err(err).Msg("error unmarshal")
		return
	}
	if err = s.subsService.SendEmail(msgData); err != nil {
		s.logger.Error().Err(err).Msg("failed to send email")
	}
}
