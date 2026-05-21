package repository

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"jcourse/internal/domain/auth"
)

type LoginAttemptRepository struct {
	client  *redis.Client
	lockout time.Duration
}

func NewLoginAttemptRepository(client *redis.Client, lockout time.Duration) *LoginAttemptRepository {
	return &LoginAttemptRepository{client: client, lockout: lockout}
}

func (r *LoginAttemptRepository) Increment(ctx context.Context, email string) (int, error) {
	key := r.key(email)
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 && r.lockout > 0 {
		_ = r.client.Expire(ctx, key, r.lockout)
	}
	return int(count), nil
}

func (r *LoginAttemptRepository) Get(ctx context.Context, email string) (int, error) {
	val, err := r.client.Get(ctx, r.key(email)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (r *LoginAttemptRepository) Reset(ctx context.Context, email string) error {
	return r.client.Del(ctx, r.key(email)).Err()
}

func (r *LoginAttemptRepository) key(email string) string {
	return "auth:login_attempts:" + strings.ToLower(email)
}

var _ auth.LoginAttemptRepository = (*LoginAttemptRepository)(nil)
