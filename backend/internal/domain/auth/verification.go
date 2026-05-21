package auth

import (
	"context"
	"time"
)

type VerificationCode struct {
	Email     string
	Code      string
	ExpiresAt time.Time
}

type VerificationCodeRepository interface {
	ReserveSend(ctx context.Context, email string, interval time.Duration) (time.Duration, error)
	Save(ctx context.Context, code VerificationCode, ttl time.Duration) error
	Get(ctx context.Context, email string) (*VerificationCode, error)
	Delete(ctx context.Context, email string) error
}

type VerificationCodeSender interface {
	SendVerificationCode(ctx context.Context, email string, code string) error
}
