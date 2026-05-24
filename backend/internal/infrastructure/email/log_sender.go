package email

import (
	"context"
	"log"

	"jcourse/internal/domain/account"
)

type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) SendEmail(_ context.Context, email account.Email) error {
	log.Printf("email to %s, subject: %s, body: %s", email.To, email.Subject, email.Body)
	return nil
}

var _ account.EmailSender = (*LogSender)(nil)
