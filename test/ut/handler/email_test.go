package handler_test

import (
	"errors"
	"testing"

	"10.1.20.130/dropping/notification-service/internal/domain/handler"
	"10.1.20.130/dropping/notification-service/test/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type EmailHandlerSuite struct {
	suite.Suite
	subscriberHandler handler.SubscriberHandler
	subscriberService *mocks.SubscriberServiceMock
}

func (e *EmailHandlerSuite) SetupSuite() {
	e.subscriberService = new(mocks.SubscriberServiceMock)
	logger := zerolog.Nop()
	e.subscriberHandler = handler.NewSubscriberHandler(e.subscriberService, logger)
}

func (e *EmailHandlerSuite) SetupTest() {
	e.subscriberService.ExpectedCalls = nil
	e.subscriberService.Calls = nil
}

func TestSendEMailServiceSuite(t *testing.T) {
	suite.Run(t, &EmailHandlerSuite{})
}

func (e *EmailHandlerSuite) TestSubscriberHandler_Email_Success() {
	msg := &mocks.MockJetstreamMsg{Datainternal: []byte(`{"receiver":["user@example.com"],"message_type":"welcome","message":"Hello!"}`)}

	e.subscriberService.On("SendEmail", mock.Anything).Return(nil)

	err := e.subscriberHandler.EmailHandler(msg)
	e.NoError(err)
	e.subscriberService.AssertCalled(e.T(), "SendEmail", mock.Anything)
}

func (e *EmailHandlerSuite) TestSubscriberHandler_Email_UnmarshalError() {
	msg := &mocks.MockJetstreamMsg{Datainternal: []byte(`"message_type":"welcome","message":"Hello!"}`)}

	err := e.subscriberHandler.EmailHandler(msg)
	e.Error(err)
}

func (e *EmailHandlerSuite) TestSubscriberHandler_Email_UnsupportedType() {
	msg := &mocks.MockJetstreamMsg{Datainternal: []byte(`{"receiver":["user@example.com"],"message_type":"unsupported","message":"Hello!"}`)}

	e.subscriberService.On("SendEmail", mock.Anything).Return(errors.New("type not supported"))

	err := e.subscriberHandler.EmailHandler(msg)
	e.Error(err)
	e.subscriberService.AssertCalled(e.T(), "SendEmail", mock.Anything)
}
