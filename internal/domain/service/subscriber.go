package service

import (
	"errors"

	"github.com/micros-template/notification-service/internal/infrastructure/mail"
	"github.com/micros-template/sharedlib/dto"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type (
	SubscriberService interface {
		SendEmail(msg dto.MailNotificationMessage) error
	}
	subscriberService struct {
		logger zerolog.Logger
		mail   mail.Mail
	}
)

func NewSubscriberService(logger zerolog.Logger, mail mail.Mail) SubscriberService {
	return &subscriberService{
		logger: logger,
		mail:   mail,
	}
}

func (s *subscriberService) SendEmail(msg dto.MailNotificationMessage) error {
	s.mail.SetSender(viper.GetString("mail.sender"))
	s.mail.SetReceiver(msg.Receiver...)

	switch msg.MsgType {
	case "welcome":
		s.mail.SetSubject("Welcome to Dropboks!!")
		if err := s.mail.SetBody("welcome.html", struct {
			Email string
		}{
			Email: msg.Receiver[0],
		}); err != nil {
			s.logger.Error().Err(err).Msg("error set body html")
			return err
		}
	case "OTP":
		s.mail.SetSubject("OTP")
		if err := s.mail.SetBody("otp.html", struct {
			OTP string
		}{
			OTP: msg.Message,
		}); err != nil {
			s.logger.Error().Err(err).Msg("error set body html")
			return err
		}
	case "verification":
		s.mail.SetSubject("Email Verification")
		if err := s.mail.SetBody("verification.html", struct {
			LINK string
		}{
			LINK: msg.Message,
		}); err != nil {
			s.logger.Error().Err(err).Msg("error set body html")
			return err
		}
	case "changeEmail":
		s.mail.SetSubject("Change Linked Email Verification")
		if err := s.mail.SetBody("verification.html", struct {
			LINK string
		}{
			LINK: msg.Message,
		}); err != nil {
			s.logger.Error().Err(err).Msg("error set body html")
			return err
		}
	case "resetPassword":
		s.mail.SetSubject("Reset Password")
		if err := s.mail.SetBody("reset-password.html", struct {
			LINK string
		}{
			LINK: msg.Message,
		}); err != nil {
			s.logger.Error().Err(err).Msg("error set body html")
			return err
		}
	default:
		return errors.New("type not supported")
	}

	if err := s.mail.Send(); err != nil {
		return err
	}
	return nil
}
