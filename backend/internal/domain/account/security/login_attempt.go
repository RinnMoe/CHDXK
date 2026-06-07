package security

import (
	"context"
	"time"

	"jcourse/pkg/apperr"
)

var (
	ErrInvalidCredentials = apperr.ErrInvalidCredentials
	ErrLoginLocked        = apperr.ErrLoginLocked
	ErrPasswordNotSet     = apperr.ErrPasswordNotSet
)

type LoginAttemptRepository interface {
	Increment(ctx context.Context, email string, lockout time.Duration) (int, error)
	Get(ctx context.Context, email string) (int, error)
	Reset(ctx context.Context, email string) error
}
