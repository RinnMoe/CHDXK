package application

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
)

func TestAuthResolutionServiceResolveUserAPIKey(t *testing.T) {
	apiKeyRepo := &authResolutionAPIKeyRepo{
		MockApiKeyRepository: auth.MockApiKeyRepository{Key: &auth.ApiKey{ID: 1, Key: "user-key", Role: auth.ApiKeyRoleUser, UserID: 7}},
	}
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{7: {ID: 7, Role: auth.RoleUser}})
	tracker := &authResolutionAccessTracker{}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, auth.DefaultApiKeyConfig), tracker)

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
	if apiKeyRepo.FindByKeyCalls != 1 {
		t.Fatalf("FindByKey calls = %d, want 1", apiKeyRepo.FindByKeyCalls)
	}
	if userRepo.FindByIDCalls != 1 {
		t.Fatalf("FindByID calls = %d, want 1", userRepo.FindByIDCalls)
	}
	if tracker.userID != 7 {
		t.Fatalf("recorded user access id = %d, want 7", tracker.userID)
	}
	if tracker.apiKeyID != 1 {
		t.Fatalf("recorded api key access id = %d, want 1", tracker.apiKeyID)
	}
}

func TestAuthResolutionServiceResolveSystemAPIKey(t *testing.T) {
	apiKeyRepo := &authResolutionAPIKeyRepo{
		MockApiKeyRepository: auth.MockApiKeyRepository{Key: &auth.ApiKey{ID: 1, Key: "system-key", Role: auth.ApiKeyRoleSystem}},
	}
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{7: {ID: 7, Role: auth.RoleUser}})
	tracker := &authResolutionAccessTracker{}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, auth.DefaultApiKeyConfig), tracker)

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
	if userRepo.FindByIDCalls != 0 {
		t.Fatalf("FindByID calls = %d, want 0", userRepo.FindByIDCalls)
	}
	if tracker.userID != 0 {
		t.Fatalf("recorded user access id = %d, want 0", tracker.userID)
	}
	if tracker.apiKeyID != 1 {
		t.Fatalf("recorded api key access id = %d, want 1", tracker.apiKeyID)
	}
}

func TestAuthResolutionServiceResolveSessionUser(t *testing.T) {
	apiKeyRepo := &authResolutionAPIKeyRepo{}
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{7: {ID: 7, Role: auth.RoleUser}})
	tracker := &authResolutionAccessTracker{}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, auth.DefaultApiKeyConfig), tracker)

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
	if apiKeyRepo.FindByKeyCalls != 0 {
		t.Fatalf("FindByKey calls = %d, want 0", apiKeyRepo.FindByKeyCalls)
	}
	if userRepo.FindByIDCalls != 1 {
		t.Fatalf("FindByID calls = %d, want 1", userRepo.FindByIDCalls)
	}
	if tracker.userID != 7 {
		t.Fatalf("recorded user access id = %d, want 7", tracker.userID)
	}
	if tracker.apiKeyID != 0 {
		t.Fatalf("recorded api key access id = %d, want 0", tracker.apiKeyID)
	}
}

func TestAuthResolutionServiceIgnoresAccessTrackerError(t *testing.T) {
	apiKeyRepo := &authResolutionAPIKeyRepo{}
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{7: {ID: 7, Role: auth.RoleUser}})
	tracker := &authResolutionAccessTracker{recordErr: context.Canceled}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, auth.DefaultApiKeyConfig), tracker)

	resolved, err := svc.Resolve(context.Background(), "", 7)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User == nil || resolved.User.ID != 7 {
		t.Fatalf("resolved user = %+v, want user 7", resolved.User)
	}
}

type authResolutionAPIKeyRepo struct {
	auth.MockApiKeyRepository
}

type authResolutionAccessTracker struct {
	userID    int
	apiKeyID  int
	recordErr error
}

func (t *authResolutionAccessTracker) RecordUserAccess(_ context.Context, userID int, _ time.Time) error {
	t.userID = userID
	return t.recordErr
}

func (t *authResolutionAccessTracker) RecordApiKeyAccess(_ context.Context, apiKeyID int, _ time.Time) error {
	t.apiKeyID = apiKeyID
	return t.recordErr
}

func (t *authResolutionAccessTracker) Flush(context.Context) (auth.AccessFlushResult, error) {
	return auth.AccessFlushResult{}, nil
}
