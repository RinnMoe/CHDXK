package auth

import (
	"context"
	"time"
)

type ApiKey struct {
	ID        int
	Name      string
	Key       string
	CreatedAt time.Time
}

type ApiKeyRepository interface {
	FindByKey(ctx context.Context, key string) (*ApiKey, error)
}

type ApiKeyService struct {
	repo ApiKeyRepository
}

func NewApiKeyService(repo ApiKeyRepository) *ApiKeyService {
	return &ApiKeyService{repo: repo}
}

func (s *ApiKeyService) ValidateKey(ctx context.Context, key string) (bool, error) {
	apiKey, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return false, err
	}
	return apiKey != nil, nil
}
