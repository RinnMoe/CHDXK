package application

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
)

func TestAuthResolutionServiceResolveUserAPIKey(t *testing.T) {
	apiKeyRepo := &authResolutionAPIKeyRepo{
		key:    "user-key",
		apiKey: &auth.ApiKey{ID: 1, Key: "user-key", Role: auth.ApiKeyRoleUser, UserID: 7},
	}
	userRepo := &authResolutionUserRepo{user: &auth.User{ID: 7, Role: auth.RoleUser}}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, auth.DefaultApiKeyConfig))

	resolved, err := svc.Resolve(context.Background(), "user-key", 0)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User == nil || resolved.User.ID != 7 {
		t.Fatalf("resolved user = %+v, want user 7", resolved.User)
	}
	if resolved.ApiKey == nil || resolved.ApiKey.ID != 1 {
		t.Fatalf("resolved api key = %+v, want key 1", resolved.ApiKey)
	}
	if apiKeyRepo.findByKeyCalls != 1 {
		t.Fatalf("FindByKey calls = %d, want 1", apiKeyRepo.findByKeyCalls)
	}
	if userRepo.findByIDCalls != 1 {
		t.Fatalf("FindByID calls = %d, want 1", userRepo.findByIDCalls)
	}
	if !apiKeyRepo.touched {
		t.Fatal("expected api key to be touched")
	}
}

func TestAuthResolutionServiceResolveSystemAPIKey(t *testing.T) {
	apiKeyRepo := &authResolutionAPIKeyRepo{
		key:    "system-key",
		apiKey: &auth.ApiKey{ID: 1, Key: "system-key", Role: auth.ApiKeyRoleSystem},
	}
	userRepo := &authResolutionUserRepo{user: &auth.User{ID: 7, Role: auth.RoleUser}}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, auth.DefaultApiKeyConfig))

	resolved, err := svc.Resolve(context.Background(), "system-key", 7)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User != nil {
		t.Fatalf("resolved user = %+v, want nil", resolved.User)
	}
	if resolved.ApiKey == nil || resolved.ApiKey.ID != 1 {
		t.Fatalf("resolved api key = %+v, want key 1", resolved.ApiKey)
	}
	if userRepo.findByIDCalls != 0 {
		t.Fatalf("FindByID calls = %d, want 0", userRepo.findByIDCalls)
	}
	if !apiKeyRepo.touched {
		t.Fatal("expected api key to be touched")
	}
}

func TestAuthResolutionServiceResolveSessionUser(t *testing.T) {
	apiKeyRepo := &authResolutionAPIKeyRepo{}
	userRepo := &authResolutionUserRepo{user: &auth.User{ID: 7, Role: auth.RoleUser}}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, auth.DefaultApiKeyConfig))

	resolved, err := svc.Resolve(context.Background(), "", 7)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User == nil || resolved.User.ID != 7 {
		t.Fatalf("resolved user = %+v, want user 7", resolved.User)
	}
	if resolved.ApiKey != nil {
		t.Fatalf("resolved api key = %+v, want nil", resolved.ApiKey)
	}
	if apiKeyRepo.findByKeyCalls != 0 {
		t.Fatalf("FindByKey calls = %d, want 0", apiKeyRepo.findByKeyCalls)
	}
	if userRepo.findByIDCalls != 1 {
		t.Fatalf("FindByID calls = %d, want 1", userRepo.findByIDCalls)
	}
}

type authResolutionUserRepo struct {
	user          *auth.User
	findByIDCalls int
}

func (r *authResolutionUserRepo) Update(context.Context, *auth.User) error { return nil }

func (r *authResolutionUserRepo) FindByID(_ context.Context, id int) (*auth.User, error) {
	r.findByIDCalls++
	if r.user == nil || r.user.ID != id {
		return nil, nil
	}
	copy := *r.user
	return &copy, nil
}

func (r *authResolutionUserRepo) FindByRole(context.Context, string) ([]auth.User, error) {
	return nil, nil
}

type authResolutionAPIKeyRepo struct {
	key            string
	apiKey         *auth.ApiKey
	findByKeyCalls int
	touched        bool
}

func (r *authResolutionAPIKeyRepo) FindByKey(_ context.Context, key string) (*auth.ApiKey, error) {
	r.findByKeyCalls++
	if r.apiKey == nil || key != r.key {
		return nil, nil
	}
	copy := *r.apiKey
	return &copy, nil
}

func (r *authResolutionAPIKeyRepo) ListByUser(context.Context, int) ([]auth.ApiKey, error) {
	return nil, nil
}

func (r *authResolutionAPIKeyRepo) CountByUser(context.Context, int) (int, error) { return 0, nil }

func (r *authResolutionAPIKeyRepo) Create(context.Context, *auth.ApiKey) error { return nil }

func (r *authResolutionAPIKeyRepo) DeleteByUser(context.Context, int, int) (bool, error) {
	return false, nil
}

func (r *authResolutionAPIKeyRepo) TouchLastUsed(context.Context, int, time.Time) error {
	r.touched = true
	return nil
}
