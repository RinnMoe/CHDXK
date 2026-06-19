package smtp

import (
	"context"
	"fmt"

	"gopkg.in/gomail.v2"

	"jcourse/internal/domain/email"
)

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type SMTPSender struct {
	dialer *gomail.Dialer
	from   string
}

func NewSMTPSender(conf SMTPConfig) *SMTPSender {
	return &SMTPSender{
		dialer: gomail.NewDialer(conf.Host, conf.Port, conf.Username, conf.Password),
		from:   conf.From,
	}
}

func (s *SMTPSender) SendEmail(_ context.Context, email email.Email) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", email.To)
	m.SetHeader("Subject", email.Subject)
	m.SetBody("text/html", email.Body)

	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

var _ email.Sender = (*SMTPSender)(nil)
