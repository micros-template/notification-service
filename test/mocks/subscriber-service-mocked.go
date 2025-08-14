package mocks

import (
	"10.1.20.130/dropping/sharedlib/dto"
	"github.com/stretchr/testify/mock"
)

type SubscriberServiceMock struct {
	mock.Mock
}

func (m *SubscriberServiceMock) SendEmail(msg dto.MailNotificationMessage) error {
	args := m.Called(msg)
	return args.Error(0)
}
