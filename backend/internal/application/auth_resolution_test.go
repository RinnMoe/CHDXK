package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
)

func TestAuthResolutionServiceResolveUserAPIKey(t *testing.T) {
	credential := testAuthResolutionCredential(1)
	apiKeyRepo := &authResolutionAPIKeyRepo{
		MockApiKeyRepository: auth.MockApiKeyRepository{Key: &auth.ApiKey{ID: credential.KeyID, SecretHash: credential.SecretHash, Role: auth.ApiKeyRoleUser, UserID: 7}},
	}
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{7: {ID: 7, Role: auth.RoleUser}})
	tracker := &authResolutionAccessTracker{}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig), tracker, testSessionAuthService(7, "password-hash"))

	resolved, err := svc.Resolve(context.Background(), credential.Key(), 0, "")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User == nil || resolved.User.ID != 7 {
		t.Fatalf("resolved user = %+v, want user 7", resolved.User)
	}
	if resolved.ApiKey == nil || resolved.ApiKey.ID != 1 {
		t.Fatalf("resolved api key = %+v, want key 1", resolved.ApiKey)
	}
	if apiKeyRepo.GetByIDCalls != 1 {
		t.Fatalf("GetByID calls = %d, want 1", apiKeyRepo.GetByIDCalls)
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
	credential := testAuthResolutionCredential(1)
	apiKeyRepo := &authResolutionAPIKeyRepo{
		MockApiKeyRepository: auth.MockApiKeyRepository{Key: &auth.ApiKey{ID: credential.KeyID, SecretHash: credential.SecretHash, Role: auth.ApiKeyRoleSystem}},
	}
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{7: {ID: 7, Role: auth.RoleUser}})
	tracker := &authResolutionAccessTracker{}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig), tracker, testSessionAuthService(7, "password-hash"))

	resolved, err := svc.Resolve(context.Background(), credential.Key(), 7, "")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User == nil || resolved.User.Role != auth.RoleSystem {
		t.Fatalf("resolved user = %+v, want system role", resolved.User)
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
	sessionAuth := testSessionAuthService(7, "password-hash")
	sessionHash, err := sessionAuth.HashForUser(context.Background(), 7)
	if err != nil {
		t.Fatalf("HashForUser: %v", err)
	}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig), tracker, sessionAuth)

	resolved, err := svc.Resolve(context.Background(), "", 7, sessionHash)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User == nil || resolved.User.ID != 7 {
		t.Fatalf("resolved user = %+v, want user 7", resolved.User)
	}
	if resolved.ApiKey != nil {
		t.Fatalf("resolved api key = %+v, want nil", resolved.ApiKey)
	}
	if apiKeyRepo.GetByIDCalls != 0 {
		t.Fatalf("GetByID calls = %d, want 0", apiKeyRepo.GetByIDCalls)
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

func TestAuthResolutionServiceRejectsInvalidSessionHash(t *testing.T) {
	apiKeyRepo := &authResolutionAPIKeyRepo{}
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{7: {ID: 7, Role: auth.RoleUser}})
	svc := NewAuthResolutionService(
		auth.NewCurrentUserService(userRepo),
		auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig),
		nil,
		testSessionAuthService(7, "new-password-hash"),
	)

	resolved, err := svc.Resolve(context.Background(), "", 7, "sha256:old-session-hash")
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("Resolve error = %v, want ErrInvalidSession", err)
	}
	if resolved.User != nil || resolved.ApiKey != nil {
		t.Fatalf("resolved = %+v, want empty", resolved)
	}
	if userRepo.FindByIDCalls != 0 {
		t.Fatalf("FindByID calls = %d, want 0", userRepo.FindByIDCalls)
	}
}

func TestAuthResolutionServiceIgnoresAccessTrackerError(t *testing.T) {
	apiKeyRepo := &authResolutionAPIKeyRepo{}
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{7: {ID: 7, Role: auth.RoleUser}})
	tracker := &authResolutionAccessTracker{recordErr: context.Canceled}
	sessionAuth := testSessionAuthService(7, "password-hash")
	sessionHash, err := sessionAuth.HashForUser(context.Background(), 7)
	if err != nil {
		t.Fatalf("HashForUser: %v", err)
	}
	svc := NewAuthResolutionService(auth.NewCurrentUserService(userRepo), auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig), tracker, sessionAuth)

	resolved, err := svc.Resolve(context.Background(), "", 7, sessionHash)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User == nil || resolved.User.ID != 7 {
		t.Fatalf("resolved user = %+v, want user 7", resolved.User)
	}
}

func testSessionAuthService(userID int, passwordHash string) *auth.SessionAuthService {
	repo := identity.NewMockRepository(nil)
	repo.PutAccount("", &identity.Account{ID: userID, PasswordHash: passwordHash})
	return auth.NewSessionAuthService(repo, "test-session-secret")
}

type authResolutionAPIKeyRepo struct {
	auth.MockApiKeyRepository
}

type authResolutionAccessTracker struct {
	userID    int
	apiKeyID  int64
	recordErr error
}

func (t *authResolutionAccessTracker) RecordUserAccess(_ context.Context, userID int, _ time.Time) error {
	t.userID = userID
	return t.recordErr
}

func (t *authResolutionAccessTracker) RecordApiKeyAccess(_ context.Context, apiKeyID int64, _ time.Time) error {
	t.apiKeyID = apiKeyID
	return t.recordErr
}

func (t *authResolutionAccessTracker) Flush(context.Context) (auth.AccessFlushResult, error) {
	return auth.AccessFlushResult{}, nil
}

func testAuthResolutionCredential(keyID int64) auth.ApiKeyCredential {
	credential := auth.ApiKeyCredential{KeyID: keyID, Secret: []byte("1234567890abcdef")}
	credential.SecretHash = credential.HashSecret()
	return credential
}
