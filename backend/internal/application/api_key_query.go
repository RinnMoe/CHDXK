package application

import (
	"context"

	"jcourse/internal/domain/auth"
)

type ApiKeyQueryService struct {
	svc *auth.ApiKeyService
}

func NewApiKeyQueryService(svc *auth.ApiKeyService) *ApiKeyQueryService {
	return &ApiKeyQueryService{svc: svc}
}

func (s *ApiKeyQueryService) ListSystemApiKeys(ctx context.Context) ([]ApiKeyDTO, error) {
	keys, err := s.svc.ListSystemKeys(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]ApiKeyDTO, 0, len(keys))
	for _, key := range keys {
		items = append(items, newApiKeyDTO(key, key.MaskedKey()))
	}
	return items, nil
}
