package security

import (
	"context"
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrLoginLocked        = errors.New("too many failed login attempts, please try again later")
)

type LoginAttemptRepository interface {
	Increment(ctx context.Context, email string) (int, error)
	Get(ctx context.Context, email string) (int, error)
	Reset(ctx context.Context, email string) error
}
