package email

import (
	"context"

	"jcourse/internal/domain/account"
	"jcourse/pkg/logx"
)

type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) SendEmail(ctx context.Context, email account.Email) error {
	logx.Info(ctx, "email sent by log sender", "to", email.To, "subject", email.Subject, "body", email.Body)
	return nil
}

var _ account.EmailSender = (*LogSender)(nil)
