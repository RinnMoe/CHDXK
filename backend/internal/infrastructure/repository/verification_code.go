package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"jcourse/internal/domain/account/verification"
)

type VerificationCodeRepository struct {
	client    *redis.Client
	keyPrefix string
}

func NewVerificationCodeRepository(client *redis.Client) *VerificationCodeRepository {
	return NewVerificationCodeRepositoryWithPrefix(client, "register")
}

func NewVerificationCodeRepositoryWithPrefix(client *redis.Client, prefix string) *VerificationCodeRepository {
	return &VerificationCodeRepository{client: client, keyPrefix: prefix}
}

func (r *VerificationCodeRepository) ReserveSend(ctx context.Context, email string, interval time.Duration) (time.Duration, error) {
	key := r.cooldownKey(email)
	ok, err := r.client.SetNX(ctx, key, "1", interval).Result()
	if err != nil {
		logCacheAccessFailure(ctx, "setnx", key, err)
		return 0, err
	}
	if ok {
		return 0, nil
	}
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		logCacheAccessFailure(ctx, "ttl", key, err)
		return 0, err
	}
	if ttl < 0 {
		return interval, nil
	}
	return ttl, nil
}

func (r *VerificationCodeRepository) Save(ctx context.Context, code verification.Code, ttl time.Duration) error {
	value := fmt.Sprintf("%s|%d", code.Code, code.ExpiresAt.Unix())
	key := r.codeKey(code.Email)
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		logCacheAccessFailure(ctx, "set", key, err)
		return err
	}
	return nil
}

func (r *VerificationCodeRepository) Get(ctx context.Context, email string) (*verification.Code, error) {
	key := r.codeKey(email)
	value, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		logCacheAccessFailure(ctx, "get", key, err)
		return nil, err
	}
	parts := strings.Split(value, "|")
	if len(parts) != 2 {
		return nil, verification.ErrCodeInvalid
	}
	expiresAtUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Unix(expiresAtUnix, 0)
	return &verification.Code{Email: email, Code: parts[0], ExpiresAt: expiresAt}, nil
}

func (r *VerificationCodeRepository) Delete(ctx context.Context, email string) error {
	key := r.codeKey(email)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		logCacheAccessFailure(ctx, "del", key, err)
		return err
	}
	return nil
}

func (r *VerificationCodeRepository) codeKey(email string) string {
	return redisKey("auth", r.keyPrefix, "code", strings.ToLower(email))
}

func (r *VerificationCodeRepository) cooldownKey(email string) string {
	return redisKey("auth", r.keyPrefix, "code_cooldown", strings.ToLower(email))
}

var _ verification.CodeRepository = (*VerificationCodeRepository)(nil)
