package repository

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"jcourse/internal/domain/account/security"
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
		logCacheAccessFailure(ctx, "incr", key, err)
		return 0, err
	}
	if count == 1 && r.lockout > 0 {
		if err := r.client.Expire(ctx, key, r.lockout).Err(); err != nil {
			logCacheAccessFailure(ctx, "expire", key, err)
		}
	}
	return int(count), nil
}

func (r *LoginAttemptRepository) Get(ctx context.Context, email string) (int, error) {
	key := r.key(email)
	val, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		logCacheAccessFailure(ctx, "get", key, err)
		return 0, err
	}
	return val, nil
}

func (r *LoginAttemptRepository) Reset(ctx context.Context, email string) error {
	key := r.key(email)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		logCacheAccessFailure(ctx, "del", key, err)
		return err
	}
	return nil
}

func (r *LoginAttemptRepository) key(email string) string {
	return redisKey("auth", "login_attempts", strings.ToLower(email))
}

var _ security.LoginAttemptRepository = (*LoginAttemptRepository)(nil)
