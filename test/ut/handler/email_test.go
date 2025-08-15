package handler_test

import (
	"errors"
	"testing"
	"time"

	"github.com/micros-template/notification-service/internal/domain/handler"
	"github.com/micros-template/notification-service/test/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type EmailHandlerSuite struct {
	suite.Suite
	subscriberHandler handler.SubscriberHandler
	subscriberService *mocks.SubscriberServiceMock
	logEmitter        *mocks.LoggerInfraMock
}

func (e *EmailHandlerSuite) SetupSuite() {
	e.subscriberService = new(mocks.SubscriberServiceMock)
	e.logEmitter = new(mocks.LoggerInfraMock)
	logger := zerolog.Nop()
	e.subscriberHandler = handler.NewSubscriberHandler(e.subscriberService, e.logEmitter, logger)
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
	e.logEmitter.On("EmitLog", mock.Anything, mock.Anything).Return(nil).Once()
	err := e.subscriberHandler.EmailHandler(msg)

	e.NoError(err)
	e.subscriberService.AssertCalled(e.T(), "SendEmail", mock.Anything)

	time.Sleep(time.Second)
	e.logEmitter.AssertExpectations(e.T())
}

func (e *EmailHandlerSuite) TestSubscriberHandler_Email_UnmarshalError() {
	msg := &mocks.MockJetstreamMsg{Datainternal: []byte(`"message_type":"welcome","message":"Hello!"}`)}
	e.logEmitter.On("EmitLog", mock.Anything, mock.Anything).Return(nil).Once()

	err := e.subscriberHandler.EmailHandler(msg)
	e.Error(err)

	time.Sleep(time.Second)
	e.logEmitter.AssertExpectations(e.T())
}

func (e *EmailHandlerSuite) TestSubscriberHandler_Email_UnsupportedType() {
	msg := &mocks.MockJetstreamMsg{Datainternal: []byte(`{"receiver":["user@example.com"],"message_type":"unsupported","message":"Hello!"}`)}

	e.logEmitter.On("EmitLog", mock.Anything, mock.Anything).Return(nil).Once()
	e.subscriberService.On("SendEmail", mock.Anything).Return(errors.New("type not supported"))

	err := e.subscriberHandler.EmailHandler(msg)
	e.Error(err)
	e.subscriberService.AssertCalled(e.T(), "SendEmail", mock.Anything)

	time.Sleep(time.Second)
	e.logEmitter.AssertExpectations(e.T())
}
