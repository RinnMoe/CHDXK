package verification

import (
	"context"
	"strings"
	"time"

	"jcourse/pkg/apperr"
)

var (
	ErrSendTooSoon = apperr.ErrVerificationSendTooSoon
	ErrCodeInvalid = apperr.ErrVerificationCodeInvalid
)

type Config struct {
	CodeInterval time.Duration
	CodeTTL      time.Duration
	CodeLength   int
}

var DefaultConfig = Config{
	CodeInterval: time.Minute,
	CodeTTL:      10 * time.Minute,
	CodeLength:   6,
}

func (c Config) WithDefaults() Config {
	defaults := DefaultConfig
	if c.CodeInterval <= 0 {
		c.CodeInterval = defaults.CodeInterval
	}
	if c.CodeTTL <= 0 {
		c.CodeTTL = defaults.CodeTTL
	}
	if c.CodeLength <= 0 {
		c.CodeLength = defaults.CodeLength
	}
	return c
}

type Code struct {
	Email     string
	Code      string
	ExpiresAt time.Time
}

func NewCode(email string, now time.Time, config Config) (Code, error) {
	config = config.WithDefaults()
	value, err := numericCode(config.CodeLength)
	if err != nil {
		return Code{}, err
	}
	return Code{Email: email, Code: value, ExpiresAt: now.Add(config.CodeTTL)}, nil
}

func (c *Code) Matches(value string, now time.Time) bool {
	if c == nil || now.After(c.ExpiresAt) {
		return false
	}
	return c.Code == strings.TrimSpace(value)
}

type CodeRepository interface {
	ReserveSend(ctx context.Context, email string, interval time.Duration) (time.Duration, error)
	Save(ctx context.Context, code Code, ttl time.Duration) error
	Get(ctx context.Context, email string) (*Code, error)
	Delete(ctx context.Context, email string) error
}
