package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
)

func TestAuthResolutionServiceResolveSystemAPIKey(t *testing.T) {
	credential := testAuthResolutionCredential(1)
	repo := &auth.MockApiKeyRepository{Key: &auth.ApiKey{
		ID:         credential.KeyID,
		SecretHash: credential.SecretHash,
		Role:       auth.ApiKeyRoleSystem,
	}}
	tracker := &authResolutionAccessTracker{}
	svc := NewAuthResolutionService(auth.NewApiKeyService(repo, repo, auth.DefaultApiKeyConfig), tracker, nil, nil)

	resolved, err := svc.Resolve(context.Background(), credential.Key(), "")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User == nil || resolved.User.Role != auth.RoleSystem {
		t.Fatalf("resolved user = %+v, want system role", resolved.User)
	}
	if resolved.ApiKey == nil || resolved.ApiKey.ID != credential.KeyID {
		t.Fatalf("resolved api key = %+v, want key %d", resolved.ApiKey, credential.KeyID)
	}
	if tracker.apiKeyID != credential.KeyID || tracker.userID != 0 {
		t.Fatalf("access record = user %d, api key %d", tracker.userID, tracker.apiKeyID)
	}
}

func TestAuthResolutionServiceRejectsNonSystemAPIKey(t *testing.T) {
	credential := testAuthResolutionCredential(1)
	repo := &auth.MockApiKeyRepository{Key: &auth.ApiKey{
		ID:         credential.KeyID,
		SecretHash: credential.SecretHash,
		Role:       "user",
	}}
	svc := NewAuthResolutionService(auth.NewApiKeyService(repo, repo, auth.DefaultApiKeyConfig), nil, nil, nil)

	resolved, err := svc.Resolve(context.Background(), credential.Key(), "")
	if !errors.Is(err, auth.ErrInvalidApiKeyRole) {
		t.Fatalf("Resolve error = %v, want ErrInvalidApiKeyRole", err)
	}
	if resolved.User != nil {
		t.Fatalf("resolved user = %+v, want nil", resolved.User)
	}
}

func TestAuthResolutionServiceResolveOIDCSession(t *testing.T) {
	apiKeyRepo := &auth.MockApiKeyRepository{}
	client := &fakeResHubOIDCClient{identity: &auth.ResHubIdentity{
		Issuer:        "https://hub.example.test",
		Subject:       "user:1",
		Username:      "alice",
		Email:         "alice@example.edu",
		EffectiveRole: "member",
	}}
	projection := &fakeIdentityProjection{user: &auth.User{ID: 7, Subject: "user:1", Role: auth.RoleUser}}
	tracker := &authResolutionAccessTracker{}
	svc := NewAuthResolutionService(
		auth.NewApiKeyService(apiKeyRepo, apiKeyRepo, auth.DefaultApiKeyConfig),
		tracker,
		client,
		projection,
	)

	resolved, err := svc.Resolve(context.Background(), "", "access-token")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User == nil || resolved.User.ID != 7 || resolved.User.Subject != "user:1" {
		t.Fatalf("resolved user = %+v", resolved.User)
	}
	if resolved.ApiKey != nil {
		t.Fatalf("resolved api key = %+v, want nil", resolved.ApiKey)
	}
	if client.userinfoToken != "access-token" || projection.identity.Subject != "user:1" {
		t.Fatalf("OIDC calls = client %#v projection %#v", client, projection)
	}
	if tracker.userID != 7 || tracker.apiKeyID != 0 {
		t.Fatalf("access record = user %d, api key %d", tracker.userID, tracker.apiKeyID)
	}
}

func TestAuthResolutionServiceWithoutCredentialsIsAnonymous(t *testing.T) {
	svc := NewAuthResolutionService(nil, nil, nil, nil)
	resolved, err := svc.Resolve(context.Background(), "", "")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.User != nil || resolved.ApiKey != nil {
		t.Fatalf("resolved = %+v, want anonymous", resolved)
	}
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
