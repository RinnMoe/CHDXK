package auth

import "context"

type LoginAttemptRepository interface {
	Increment(ctx context.Context, email string) (int, error)
	Get(ctx context.Context, email string) (int, error)
	Reset(ctx context.Context, email string) error
}
