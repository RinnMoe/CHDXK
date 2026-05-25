package application

import (
	"context"

	"jcourse/internal/domain/auth"
)

type ResolvedAuth struct {
	User   *auth.User
	ApiKey *auth.ApiKey
}

type AuthResolutionService struct {
	currentUserSvc *auth.AuthUserService
	apiKeySvc      *auth.ApiKeyService
}

func NewAuthResolutionService(currentUserSvc *auth.AuthUserService, apiKeySvc *auth.ApiKeyService) *AuthResolutionService {
	return &AuthResolutionService{currentUserSvc: currentUserSvc, apiKeySvc: apiKeySvc}
}

func (s *AuthResolutionService) Resolve(ctx context.Context, bearerToken string, sessionUserID int) (*ResolvedAuth, error) {
	if bearerToken != "" {
		return s.resolveAPIKey(ctx, bearerToken)
	}
	if sessionUserID > 0 {
		return s.resolveSessionUser(ctx, sessionUserID)
	}
	return &ResolvedAuth{}, nil
}

func (s *AuthResolutionService) resolveAPIKey(ctx context.Context, token string) (*ResolvedAuth, error) {
	apiKey, err := s.apiKeySvc.ValidateKey(ctx, token)
	if err != nil || apiKey == nil {
		return &ResolvedAuth{}, err
	}

	resolved := &ResolvedAuth{ApiKey: apiKey}
	if apiKey.IsUser() {
		user, err := s.currentUserSvc.GetUser(ctx, apiKey.UserID)
		if err != nil || user == nil {
			return resolved, err
		}
		resolved.User = user
	}
	if err := s.apiKeySvc.MarkKeyUsed(ctx, apiKey.ID); err != nil {
		return nil, err
	}
	return resolved, nil
}

func (s *AuthResolutionService) resolveSessionUser(ctx context.Context, userID int) (*ResolvedAuth, error) {
	user, err := s.currentUserSvc.GetUser(ctx, userID)
	if err != nil || user == nil {
		return &ResolvedAuth{}, err
	}
	return &ResolvedAuth{User: user}, nil
}
