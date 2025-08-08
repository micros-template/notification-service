package it

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	_mq "10.1.20.130/dropping/notification-service/internal/infrastructure/message-queue"
	"10.1.20.130/dropping/notification-service/test/helper"
	"github.com/dropboks/sharedlib/dto"
	_helper "github.com/dropboks/sharedlib/test/helper"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type EmailITSuite struct {
	suite.Suite
	ctx context.Context

	network                      *testcontainers.DockerNetwork
	jsConnection                 _mq.Nats
	natsContainer                *_helper.NatsContainer
	notificationServiceContainer *_helper.NotificationServiceContainer
	mailHogContainer             *_helper.MailhogContainer
}

func (e *EmailITSuite) SetupSuite() {
	log.Println("Setting up integration test suite for EmailITSuite")
	e.ctx = context.Background()

	viper.SetConfigName("config.test")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../")
	if err := viper.ReadInConfig(); err != nil {
		panic("failed to read config")
	}

	e.network = _helper.StartNetwork(e.ctx)

	// spawn nats
	nContainer, err := _helper.StartNatsContainer(e.ctx, e.network.Name, viper.GetString("container.nats_version"))
	if err != nil {
		log.Fatalf("failed starting nats container: %e", err)
	}
	e.natsContainer = nContainer

	noContainer, err := _helper.StartNotificationServiceContainer(e.ctx, e.network.Name, viper.GetString("container.notification_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting notification service container: %e", err)
	}
	e.notificationServiceContainer = noContainer

	mailContainer, err := _helper.StartMailhogContainer(e.ctx, e.network.Name, viper.GetString("container.mailhog_version"))
	if err != nil {
		log.Fatalf("failed starting mailhog container: %e", err)
	}
	e.mailHogContainer = mailContainer

	addr := fmt.Sprintf("%s://%s:%s", viper.GetString("nats.protocol"), viper.GetString("nats.test_address"), viper.GetString("nats.port"))
	nc, err := nats.Connect(addr,
		nats.UserInfo(viper.GetString("nats.credential.user"), viper.GetString("nats.credential.password")),
		nats.Name(viper.GetString("nats.connetion_name")),
		nats.Timeout(viper.GetDuration("nats.timeout")*time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(viper.GetDuration("nats.timeout")*time.Second),
	)
	if err != nil {
		panic("failed to connect to nats server")
	}

	logger := zerolog.Nop()
	jetstreamCon, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("failed starting jetstream con: %e", err)
	}
	natsInfra := _mq.NewNatsInfrastructure(nc, logger, jetstreamCon)
	err = natsInfra.CreateOrUpdateNewStream(e.ctx, &jetstream.StreamConfig{
		Name:        viper.GetString("jetstream.notification.stream.name"),
		Description: viper.GetString("jetstream.notification.stream.description"),
		Subjects:    []string{viper.GetString("jetstream.notification.subject.global")},
		MaxBytes:    10 * 1024 * 1024,
		Storage:     jetstream.FileStorage,
	})
	if err != nil {
		log.Fatal(err)
	}
	e.jsConnection = natsInfra
}

func (e *EmailITSuite) TearDownSuite() {
	if err := e.natsContainer.Terminate(e.ctx); err != nil {
		log.Fatalf("error terminating nats container: %e", err)
	}
	if err := e.notificationServiceContainer.Terminate(e.ctx); err != nil {
		log.Fatalf("error terminating notification service container: %e", err)
	}
	if err := e.mailHogContainer.Terminate(e.ctx); err != nil {
		log.Fatalf("error terminating mailhog container: %e", err)
	}
	log.Println("Tear Down integration test suite for EmailITSuite")

}
func TestEmailITSuite(t *testing.T) {
	suite.Run(t, &EmailITSuite{})
}

func (e *EmailITSuite) TestEmailIT_Success() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	subject := fmt.Sprintf("%s.%s", viper.GetString("jetstream.notification.client.mail"), "random-id")
	msg := dto.MailNotificationMessage{
		Receiver: []string{email},
		MsgType:  "OTP",
		Message:  "123456",
	}
	marshalledMsg, err := json.Marshal(msg)
	e.NoError(err)

	_, err = e.jsConnection.Publish(e.ctx, subject, []byte(marshalledMsg))
	e.NoError(err)

	regex := `<div class="otp">\s*([0-9]{4,8})\s*</div>`
	otp := helper.RetrieveDataFromEmail(email, regex, "otp", e.T())
	e.NotEmpty(otp)
}

func (e *EmailITSuite) TestEmailIT_ErrorUnmarshal() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	subject := fmt.Sprintf("%s.%s", viper.GetString("jetstream.notification.client.mail"), "random-id")
	msg := dto.MailNotificationMessage{
		Receiver: []string{email},
		Message:  "123456",
	}
	marshalledMsg, err := json.Marshal(msg)
	e.NoError(err)

	_, err = e.jsConnection.Publish(e.ctx, subject, []byte(marshalledMsg))
	e.NoError(err)

	regex := `<div class="otp">\s*([0-9]{4,8})\s*</div>`
	otp := helper.RetrieveDataFromEmail(email, regex, "otp", e.T())
	e.Empty(otp)
}

func (e *EmailITSuite) TestEmailIT_UnsupportedType() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	subject := fmt.Sprintf("%s.%s", viper.GetString("jetstream.notification.client.mail"), "random-id")
	msg := dto.MailNotificationMessage{
		Receiver: []string{email},
		MsgType:  "not-supported-email-type",
		Message:  "123456",
	}
	marshalledMsg, err := json.Marshal(msg)
	e.NoError(err)

	_, err = e.jsConnection.Publish(e.ctx, subject, []byte(marshalledMsg))
	e.NoError(err)

	regex := `<div class="otp">\s*([0-9]{4,8})\s*</div>`
	otp := helper.RetrieveDataFromEmail(email, regex, "otp", e.T())
	e.Empty(otp)
}
