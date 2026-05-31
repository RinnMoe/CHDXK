package application

import (
	"context"
	"time"

	"jcourse/internal/domain/auth"
)

type ResolvedAuth struct {
	User   *auth.User
	ApiKey *auth.ApiKey
}

type AuthResolutionService struct {
	currentUserSvc *auth.AuthUserService
	apiKeySvc      *auth.ApiKeyService
	accessTracker  auth.AccessTracker
	sessionAuth    *auth.SessionAuthService
}

func NewAuthResolutionService(currentUserSvc *auth.AuthUserService, apiKeySvc *auth.ApiKeyService, accessTracker auth.AccessTracker, sessionAuth *auth.SessionAuthService) *AuthResolutionService {
	return &AuthResolutionService{currentUserSvc: currentUserSvc, apiKeySvc: apiKeySvc, accessTracker: accessTracker, sessionAuth: sessionAuth}
}

func (s *AuthResolutionService) Resolve(ctx context.Context, bearerToken string, sessionUserID int, sessionAuthHash string) (*ResolvedAuth, error) {
	if bearerToken != "" {
		return s.resolveAPIKey(ctx, bearerToken)
	}
	if sessionUserID > 0 {
		return s.resolveSessionUser(ctx, sessionUserID, sessionAuthHash)
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
	} else if apiKey.IsSystem() {
		resolved.User = &auth.User{Role: auth.RoleSystem}
	}
	s.recordAccess(ctx, resolved)
	return resolved, nil
}

func (s *AuthResolutionService) resolveSessionUser(ctx context.Context, userID int, sessionAuthHash string) (*ResolvedAuth, error) {
	if err := s.sessionAuth.Validate(ctx, userID, sessionAuthHash); err != nil {
		return &ResolvedAuth{}, err
	}
	user, err := s.currentUserSvc.GetUser(ctx, userID)
	if err != nil || user == nil {
		return &ResolvedAuth{}, err
	}
	resolved := &ResolvedAuth{User: user}
	s.recordAccess(ctx, resolved)
	return resolved, nil
}

func (s *AuthResolutionService) recordAccess(ctx context.Context, resolved *ResolvedAuth) {
	if s.accessTracker == nil || resolved == nil {
		return
	}
	now := time.Now()
	if resolved.ApiKey != nil {
		_ = s.accessTracker.RecordApiKeyAccess(ctx, resolved.ApiKey.ID, now)
	}
	if resolved.User != nil {
		if resolved.User.IsSystemAPIKey() {
			return
		}
		_ = s.accessTracker.RecordUserAccess(ctx, resolved.User.ID, now)
	}
}
