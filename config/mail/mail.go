package mail

import (
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

func New() (*gomail.Message, *gomail.Dialer) {
	host := viper.GetString("mail.host")
	port := viper.GetInt("mail.port")
	username := viper.GetString("mail.username")
	password := viper.GetString("mail.password")

	dialer := gomail.NewDialer(host, port, username, password)
	return gomail.NewMessage(), dialer
}
