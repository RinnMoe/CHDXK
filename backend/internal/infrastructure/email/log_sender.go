package email

import (
	"context"
	"log"

	"jcourse/internal/domain/account"
)

type LogVerificationCodeSender struct{}

func NewLogVerificationCodeSender() *LogVerificationCodeSender {
	return &LogVerificationCodeSender{}
}

func (s *LogVerificationCodeSender) SendVerificationCode(_ context.Context, email string, code string) error {
	log.Printf("verification code for %s: %s", email, code)
	return nil
}

var _ account.VerificationCodeSender = (*LogVerificationCodeSender)(nil)
