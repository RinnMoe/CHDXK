package email

import (
	"context"

	domainemail "jcourse/internal/domain/email"
	"jcourse/pkg/logx"
)

type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) SendEmail(ctx context.Context, email domainemail.Email) error {
	logx.Info(ctx, "email sent by log sender", "to", email.To, "subject", email.Subject, "body", email.Body)
	return nil
}

var _ domainemail.Sender = (*LogSender)(nil)
