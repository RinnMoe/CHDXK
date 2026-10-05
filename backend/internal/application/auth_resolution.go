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
	bootstrap     auth.BootstrapAccountRepository
	apiKeySvc     *auth.ApiKeyService
	accessTracker auth.AccessTracker
	reshub        auth.ResHubOIDCClient
	projection    auth.IdentityProjectionRepository
}

func NewAuthResolutionService(apiKeySvc *auth.ApiKeyService, accessTracker auth.AccessTracker, reshub auth.ResHubOIDCClient, projection auth.IdentityProjectionRepository) *AuthResolutionService {
	return &AuthResolutionService{
		apiKeySvc:     apiKeySvc,
		accessTracker: accessTracker,
		reshub:        reshub,
		projection:    projection,
	}
}

func (s *AuthResolutionService) Resolve(ctx context.Context, bearerToken, oidcAccessToken string) (*ResolvedAuth, error) {
	if bearerToken != "" {
		return s.resolveAPIKey(ctx, bearerToken)
	}
	if oidcAccessToken == "" {
		return &ResolvedAuth{}, nil
	}
	if s.reshub == nil || s.projection == nil {
		return &ResolvedAuth{}, auth.ErrResHubUnavailable
	}
	return s.resolveResHubSession(ctx, oidcAccessToken)
}

func (s *AuthResolutionService) resolveAPIKey(ctx context.Context, token string) (*ResolvedAuth, error) {
	apiKey, err := s.apiKeySvc.ValidateKey(ctx, token)
	if err != nil || apiKey == nil {
		return &ResolvedAuth{}, err
	}

	resolved := &ResolvedAuth{ApiKey: apiKey}
	if apiKey.IsSystem() {
		resolved.User = &auth.User{Role: auth.RoleSystem}
	} else {
		return resolved, auth.ErrInvalidApiKeyRole
	}
	s.recordAccess(ctx, resolved)
	return resolved, nil
}

func (s *AuthResolutionService) resolveResHubSession(ctx context.Context, token string) (*ResolvedAuth, error) {
	snapshot, err := s.reshub.UserInfo(ctx, token)
	if err != nil {
		return &ResolvedAuth{}, err
	}
	user, err := s.projection.Upsert(ctx, snapshot)
	if err != nil {
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

func (s *AuthResolutionService) SetBootstrap(repo auth.BootstrapAccountRepository) {
	s.bootstrap = repo
}
func (s *AuthResolutionService) ResolveAccount(ctx context.Context, bearer, token string, bootstrap bool) (*ResolvedAuth, error) {
	if !bootstrap || bearer != "" {
		return s.Resolve(ctx, bearer, token)
	}
	if s.bootstrap == nil {
		return nil, auth.ErrResHubSession
	}
	if token != "" {
		if s.reshub == nil {
			return nil, auth.ErrResHubUnavailable
		}
		snapshot, err := s.reshub.UserInfo(ctx, token)
		if err != nil {
			return nil, err
		}
		ok, err := s.bootstrap.Matches(ctx, snapshot.Issuer, snapshot.Subject)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, auth.ErrResHubSession
		}
	}
	user, err := s.bootstrap.Current(ctx)
	if err != nil {
		return nil, err
	}
	resolved := &ResolvedAuth{User: user}
	s.recordAccess(ctx, resolved)
	return resolved, nil
}
