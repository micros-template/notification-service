package service_test

import (
	"testing"

	"github.com/micros-template/notification-service/internal/domain/service"
	"github.com/micros-template/notification-service/test/mocks"
	"github.com/micros-template/sharedlib/dto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SendEmailServiceSuite struct {
	suite.Suite
	subscriberService service.SubscriberService
	mail              *mocks.MailMock
}

func (s *SendEmailServiceSuite) SetupSuite() {
	s.mail = new(mocks.MailMock)
	logger := zerolog.Nop()
	s.subscriberService = service.NewSubscriberService(logger, s.mail)
}

func (s *SendEmailServiceSuite) SetupTest() {
	s.mail.ExpectedCalls = nil
	s.mail.Calls = nil
}

func TestSendEMailServiceSuite(t *testing.T) {
	suite.Run(t, &SendEmailServiceSuite{})
}
func (s *SendEmailServiceSuite) TestSubscriberService_SendEmail_Success() {
	msg := dto.MailNotificationMessage{
		Receiver: []string{"random@mail.com"},
		MsgType:  "welcome",
		Message:  "random@mail.com",
	}
	s.mail.On("SetSender", mock.Anything)
	s.mail.On("SetReceiver", mock.Anything)
	s.mail.On("SetSubject", mock.Anything)
	s.mail.On("SetBody", "welcome.html", mock.Anything).Return(nil)
	s.mail.On("Send").Return(nil)
	err := s.subscriberService.SendEmail(msg)
	s.NoError(err)
	s.mail.AssertExpectations(s.T())
}

func (s *SendEmailServiceSuite) TestSubscriberService_SendEmail_TypeNotSupported() {
	msg := dto.MailNotificationMessage{
		Receiver: []string{"random@mail.com"},
		MsgType:  "shouldbeerror",
		Message:  "random@mail.com",
	}

	s.mail.On("SetSender", mock.Anything)
	s.mail.On("SetReceiver", mock.Anything)

	err := s.subscriberService.SendEmail(msg)
	s.Error(err)
	s.mail.AssertExpectations(s.T())
}
