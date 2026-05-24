package account

import (
	"context"
	"crypto/subtle"
	"strings"
	"time"
)

type VerificationCode struct {
	Email     string
	Code      string
	ExpiresAt time.Time
}

func (c *VerificationCode) Matches(value string, now time.Time) bool {
	if c == nil || now.After(c.ExpiresAt) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(c.Code), []byte(strings.TrimSpace(value))) == 1
}

type VerificationCodeRepository interface {
	ReserveSend(ctx context.Context, email string, interval time.Duration) (time.Duration, error)
	Save(ctx context.Context, code VerificationCode, ttl time.Duration) error
	Get(ctx context.Context, email string) (*VerificationCode, error)
	Delete(ctx context.Context, email string) error
}
