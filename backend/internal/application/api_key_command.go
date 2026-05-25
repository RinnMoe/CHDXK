package application

import (
	"context"

	"jcourse/internal/domain/auth"
)

type ApiKeyCommandService struct {
	svc *auth.ApiKeyService
}

func NewApiKeyCommandService(svc *auth.ApiKeyService) *ApiKeyCommandService {
	return &ApiKeyCommandService{svc: svc}
}

func (s *ApiKeyCommandService) CreateMyApiKey(ctx context.Context, userID int, cmd CreateApiKeyCommand) (*ApiKeyDTO, error) {
	key, err := s.svc.CreateUserKey(ctx, userID, cmd.Name)
	if err != nil {
		return nil, err
	}
	dto := newApiKeyDTO(*key, true)
	return &dto, nil
}

func (s *ApiKeyCommandService) DeleteMyApiKey(ctx context.Context, userID int, id int) error {
	return s.svc.DeleteUserKey(ctx, userID, id)
}

func (s *ApiKeyCommandService) CreateSystemApiKey(ctx context.Context, cmd CreateApiKeyCommand) (*ApiKeyDTO, error) {
	key, err := s.svc.CreateSystemKey(ctx, cmd.Name)
	if err != nil {
		return nil, err
	}
	dto := newApiKeyDTO(*key, true)
	return &dto, nil
}

func (s *ApiKeyCommandService) DeleteSystemApiKey(ctx context.Context, id int) error {
	return s.svc.DeleteSystemKey(ctx, id)
}
