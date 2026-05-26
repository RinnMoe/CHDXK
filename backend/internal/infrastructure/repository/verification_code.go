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
		return 0, err
	}
	if ok {
		return 0, nil
	}
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if ttl < 0 {
		return interval, nil
	}
	return ttl, nil
}

func (r *VerificationCodeRepository) Save(ctx context.Context, code verification.Code, ttl time.Duration) error {
	value := fmt.Sprintf("%s|%d", code.Code, code.ExpiresAt.Unix())
	return r.client.Set(ctx, r.codeKey(code.Email), value, ttl).Err()
}

func (r *VerificationCodeRepository) Get(ctx context.Context, email string) (*verification.Code, error) {
	value, err := r.client.Get(ctx, r.codeKey(email)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
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
	return r.client.Del(ctx, r.codeKey(email)).Err()
}

func (r *VerificationCodeRepository) codeKey(email string) string {
	return redisKey("auth", r.keyPrefix, "code", strings.ToLower(email))
}

func (r *VerificationCodeRepository) cooldownKey(email string) string {
	return redisKey("auth", r.keyPrefix, "code_cooldown", strings.ToLower(email))
}

var _ verification.CodeRepository = (*VerificationCodeRepository)(nil)
