package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MailMock struct {
	mock.Mock
}

func (m *MailMock) SetSender(sender string) {
	m.Called(sender)
}

func (m *MailMock) SetReceiver(to ...string) {
	m.Called(to)
}

func (m *MailMock) SetSubject(subject string) {
	m.Called(subject)
}

func (m *MailMock) SetBody(templateFileName string, data interface{}) error {
	args := m.Called(templateFileName, data)
	return args.Error(0)
}

func (m *MailMock) Send() error {
	args := m.Called()
	return args.Error(0)
}
