package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

const (
	ApiKeyRoleSystem = "system"
	ApiKeyRoleUser   = "user"
	apiKeyPrefix     = "jc_"
)

type ApiKey struct {
	ID         int
	Name       string
	Key        string
	Role       string
	UserID     int
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

type ApiKeyRepository interface {
	FindByKey(ctx context.Context, key string) (*ApiKey, error)
	ListByUser(ctx context.Context, userID int) ([]ApiKey, error)
	Create(ctx context.Context, apiKey *ApiKey) error
	DeleteByUser(ctx context.Context, id int, userID int) (bool, error)
	TouchLastUsed(ctx context.Context, id int, at time.Time) error
}

type ApiKeyService struct {
	repo ApiKeyRepository
}

func NewApiKeyService(repo ApiKeyRepository) *ApiKeyService {
	return &ApiKeyService{repo: repo}
}

func NewUserApiKey(name string, userID int, now time.Time) (*ApiKey, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrApiKeyNameRequired
	}
	key, err := GenerateApiKey()
	if err != nil {
		return nil, err
	}
	return &ApiKey{
		Name:      name,
		Key:       key,
		Role:      ApiKeyRoleUser,
		UserID:    userID,
		CreatedAt: now,
	}, nil
}

func GenerateApiKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return apiKeyPrefix + hex.EncodeToString(b), nil
}

func (k ApiKey) IsSystem() bool {
	return k.Role == ApiKeyRoleSystem && k.UserID == 0
}

func (k ApiKey) IsUser() bool {
	return k.Role == ApiKeyRoleUser && k.UserID > 0
}

func (k ApiKey) BelongsTo(userID int) bool {
	return k.IsUser() && k.UserID == userID
}

func MaskApiKey(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:7] + strings.Repeat("*", 12) + key[len(key)-6:]
}

func (s *ApiKeyService) ValidateKey(ctx context.Context, key string) (*ApiKey, error) {
	apiKey, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, nil
	}
	return apiKey, nil
}

func (s *ApiKeyService) MarkKeyUsed(ctx context.Context, id int) error {
	return s.repo.TouchLastUsed(ctx, id, time.Now())
}

func (s *ApiKeyService) ListUserKeys(ctx context.Context, userID int) ([]ApiKey, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *ApiKeyService) CreateUserKey(ctx context.Context, userID int, name string) (*ApiKey, error) {
	apiKey, err := NewUserApiKey(name, userID, time.Now())
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, apiKey); err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (s *ApiKeyService) DeleteUserKey(ctx context.Context, userID int, id int) error {
	deleted, err := s.repo.DeleteByUser(ctx, id, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrApiKeyNotFound
	}
	return nil
}
