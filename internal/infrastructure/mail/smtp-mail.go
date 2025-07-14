package mail

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

type (
	Mail interface {
		SetSender(sender string)
		SetReceiver(to ...string)
		SetSubject(subject string)
		SetBody(templateFileName string, data interface{}) error
		Send() error
	}
	mail struct {
		msg         *gomail.Message
		dialer      *gomail.Dialer
		htmlRootDir string
	}
)

func New(msg *gomail.Message, dialer *gomail.Dialer) Mail {
	return &mail{
		msg:         msg,
		dialer:      dialer,
		htmlRootDir: viper.GetString("mail.html_root_dir"),
	}
}

func (m *mail) SetSender(sender string) {
	m.msg.SetHeader("From", sender)
}

func (m *mail) SetReceiver(to ...string) {
	m.msg.SetHeader("To", to...)
}

func (m *mail) SetSubject(subject string) {
	m.msg.SetHeader("Subject", subject)
}

func (m *mail) SetBody(templateFileName string, data interface{}) error {
	file := fmt.Sprintf("%s/%s", m.htmlRootDir, templateFileName)
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		return err
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}
	m.msg.SetBody("text/html", body.String())
	return nil
}

func (m *mail) Send() error {
	if err := m.dialer.DialAndSend(m.msg); err != nil {
		return err
	}
	return nil
}
