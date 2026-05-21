package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"jcourse/internal/domain/auth"
)

type VerificationCodeRepository struct {
	client *redis.Client
}

func NewVerificationCodeRepository(client *redis.Client) *VerificationCodeRepository {
	return &VerificationCodeRepository{client: client}
}

func (r *VerificationCodeRepository) ReserveSend(ctx context.Context, email string, interval time.Duration) (time.Duration, error) {
	key := verificationCooldownKey(email)
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

func (r *VerificationCodeRepository) Save(ctx context.Context, code auth.VerificationCode, ttl time.Duration) error {
	value := fmt.Sprintf("%s|%d", code.Code, code.ExpiresAt.Unix())
	return r.client.Set(ctx, verificationCodeKey(code.Email), value, ttl).Err()
}

func (r *VerificationCodeRepository) Get(ctx context.Context, email string) (*auth.VerificationCode, error) {
	value, err := r.client.Get(ctx, verificationCodeKey(email)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	parts := strings.Split(value, "|")
	if len(parts) != 2 {
		return nil, auth.ErrVerificationCodeInvalid
	}
	expiresAtUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Unix(expiresAtUnix, 0)
	return &auth.VerificationCode{Email: email, Code: parts[0], ExpiresAt: expiresAt}, nil
}

func (r *VerificationCodeRepository) Delete(ctx context.Context, email string) error {
	return r.client.Del(ctx, verificationCodeKey(email)).Err()
}

func verificationCodeKey(email string) string {
	return "auth:register_code:" + strings.ToLower(email)
}

func verificationCooldownKey(email string) string {
	return "auth:register_code_cooldown:" + strings.ToLower(email)
}

var _ auth.VerificationCodeRepository = (*VerificationCodeRepository)(nil)
