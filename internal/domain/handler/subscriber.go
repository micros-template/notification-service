package handler

import (
	"encoding/json"
	"fmt"

	"10.1.20.130/dropping/notification-service/internal/domain/service"
	"10.1.20.130/dropping/notification-service/internal/infrastructure/logger"
	"10.1.20.130/dropping/sharedlib/dto"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
)

type (
	SubscriberHandler interface {
		EmailHandler(msg jetstream.Msg) error
	}
	subscriberHandler struct {
		subsService service.SubscriberService
		logger      zerolog.Logger
		logEmitter  logger.LoggerInfra
	}
)

func NewSubscriberHandler(svc service.SubscriberService, logEmitter logger.LoggerInfra, logger zerolog.Logger) SubscriberHandler {
	return &subscriberHandler{
		subsService: svc,
		logEmitter:  logEmitter,
		logger:      logger,
	}
}

func (s *subscriberHandler) EmailHandler(msg jetstream.Msg) error {
	var msgData dto.MailNotificationMessage
	err := json.Unmarshal(msg.Data(), &msgData)
	if err != nil {
		go func() {
			if err := s.logEmitter.EmitLog("ERR", fmt.Sprintf("error unmarshal: %v", err.Error())); err != nil {
				s.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return err
	}
	if err = s.subsService.SendEmail(msgData); err != nil {
		go func() {
			if err := s.logEmitter.EmitLog("ERR", fmt.Sprintf("Failed to send email. type: %s receiver: %s Err: %v ", msgData.MsgType, msgData.Receiver, err)); err != nil {
				s.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return err
	}
	go func() {
		if err := s.logEmitter.EmitLog("INFO", fmt.Sprintf("Email sent. type: %s receiver: %s ", msgData.MsgType, msgData.Receiver)); err != nil {
			s.logger.Error().Err(err).Msg("failed to emit log")
		}
	}()
	return nil
}
