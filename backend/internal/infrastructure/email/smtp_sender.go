package email

import (
	"context"
	"fmt"

	"gopkg.in/gomail.v2"

	"jcourse/config"
	"jcourse/internal/domain/account"
)

type SMTPVerificationCodeSender struct {
	dialer *gomail.Dialer
	from   string
}

func NewSMTPVerificationCodeSender(conf config.SMTPConfig) *SMTPVerificationCodeSender {
	return &SMTPVerificationCodeSender{
		dialer: gomail.NewDialer(conf.Host, conf.Port, conf.Username, conf.Password),
		from:   conf.From,
	}
}

func (s *SMTPVerificationCodeSender) SendVerificationCode(_ context.Context, email string, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your verification code")
	m.SetBody("text/plain", fmt.Sprintf("Your verification code is: %s", code))

	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}
	return nil
}

var _ account.VerificationCodeSender = (*SMTPVerificationCodeSender)(nil)
