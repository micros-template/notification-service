package mocks

import (
	"context"

	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type MockJetstreamMsg struct {
	Datainternal []byte
}

func (m *MockJetstreamMsg) Data() []byte                              { return m.Datainternal }
func (m *MockJetstreamMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *MockJetstreamMsg) Headers() nats.Header                      { return nil }
func (m *MockJetstreamMsg) Subject() string                           { return "test.subject" }
func (m *MockJetstreamMsg) Reply() string                             { return "" }
func (m *MockJetstreamMsg) Ack() error                                { return nil }
func (m *MockJetstreamMsg) DoubleAck(ctx context.Context) error       { return nil }
func (m *MockJetstreamMsg) Nak() error                                { return nil }
func (m *MockJetstreamMsg) NakWithDelay(delay time.Duration) error    { return nil }
func (m *MockJetstreamMsg) InProgress() error                         { return nil }
func (m *MockJetstreamMsg) Term() error                               { return nil }
func (m *MockJetstreamMsg) TermWithReason(reason string) error        { return nil }
