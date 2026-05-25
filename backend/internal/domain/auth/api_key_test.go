package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestApiKeyCredentialKeyRoundTrip(t *testing.T) {
	credential := testApiKeyCredential(12345)
	key := credential.Key()

	parsed, err := ParseApiKeyCredential(key)
	if err != nil {
		t.Fatalf("ParseApiKeyCredential: %v", err)
	}
	if parsed.KeyID != credential.KeyID {
		t.Fatalf("KeyID = %d, want %d", parsed.KeyID, credential.KeyID)
	}
	if string(parsed.Secret) != string(credential.Secret) {
		t.Fatalf("Secret = %q, want %q", parsed.Secret, credential.Secret)
	}
	if parsed.SecretHash != credential.SecretHash {
		t.Fatalf("SecretHash = %q, want %q", parsed.SecretHash, credential.SecretHash)
	}
	if !strings.HasPrefix(key, "jc_") {
		t.Fatalf("key = %q, want jc_ prefix", key)
	}
}

func TestParseApiKeyCredentialRejectsInvalidKey(t *testing.T) {
	for _, key := range []string{"", "wrong", "jc_bad_bad", "jc_AAAAAAAAAAA_short"} {
		if _, err := ParseApiKeyCredential(key); !errors.Is(err, ErrInvalidApiKey) {
			t.Fatalf("ParseApiKeyCredential(%q) err = %v, want ErrInvalidApiKey", key, err)
		}
	}
}

func TestNewUserApiKey(t *testing.T) {
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
	credential := testApiKeyCredential(12)
	key, err := NewUserApiKey("  local script  ", 12, credential, now)
	if err != nil {
		t.Fatalf("NewUserApiKey: %v", err)
	}
	if key.ID != credential.KeyID || key.SecretHash != credential.SecretHash {
		t.Fatalf("unexpected credential fields: %+v", key)
	}
	if key.Name != "local script" {
		t.Fatalf("Name = %q, want %q", key.Name, "local script")
	}
	if !key.IsUser() || key.UserID != 12 {
		t.Fatalf("unexpected api key role/user: %+v", key)
	}
	if !key.CreatedAt.Equal(now) {
		t.Fatalf("CreatedAt = %v, want %v", key.CreatedAt, now)
	}
}

func TestNewUserApiKeyRejectsEmptyName(t *testing.T) {
	_, err := NewUserApiKey("  ", 12, testApiKeyCredential(12), time.Now())
	if !errors.Is(err, ErrApiKeyNameRequired) {
		t.Fatalf("err = %v, want ErrApiKeyNameRequired", err)
	}
}

func TestApiKeyMaskedKey(t *testing.T) {
	key := ApiKey{ID: 12345}
	if got, want := key.MaskedKey(), "jc_"+EncodeApiKeyID(12345)+"_********"; got != want {
		t.Fatalf("MaskedKey() = %q, want %q", got, want)
	}
}

func TestApiKeyRoleHelpers(t *testing.T) {
	if !(ApiKey{Role: ApiKeyRoleSystem, UserID: 0}).IsSystem() {
		t.Fatal("expected system key to be system")
	}
	if !(ApiKey{Role: ApiKeyRoleUser, UserID: 7}).IsUser() {
		t.Fatal("expected user key to be user")
	}
	if !(ApiKey{Role: ApiKeyRoleUser, UserID: 7}).BelongsTo(7) {
		t.Fatal("expected user key to belong to user 7")
	}
}

func TestApiKeyService_ValidateKeyReturnsKey(t *testing.T) {
	credential := testApiKeyCredential(1)
	repo := &MockApiKeyRepository{Key: &ApiKey{ID: credential.KeyID, SecretHash: credential.SecretHash, Role: ApiKeyRoleSystem}}
	svc := NewApiKeyService(repo, repo, DefaultApiKeyConfig)
	got, err := svc.ValidateKey(context.Background(), credential.Key())
	if err != nil {
		t.Fatalf("ValidateKey: %v", err)
	}
	if got == nil || got.ID != 1 {
		t.Fatalf("ValidateKey = %+v, want id 1", got)
	}
}

func TestApiKeyService_ValidateKeyRejectsSecretMismatch(t *testing.T) {
	credential := testApiKeyCredential(1)
	mismatch := ApiKeyCredential{KeyID: 1, Secret: []byte("abcdef1234567890")}
	mismatch.SecretHash = mismatch.HashSecret()
	repo := &MockApiKeyRepository{Key: &ApiKey{ID: credential.KeyID, SecretHash: mismatch.SecretHash, Role: ApiKeyRoleSystem}}
	svc := NewApiKeyService(repo, repo, DefaultApiKeyConfig)
	got, err := svc.ValidateKey(context.Background(), credential.Key())
	if err != nil {
		t.Fatalf("ValidateKey: %v", err)
	}
	if got != nil {
		t.Fatalf("ValidateKey = %+v, want nil", got)
	}
}

func TestApiKeyService_CreateUserKey(t *testing.T) {
	repo := &MockApiKeyRepository{Count: DefaultApiKeyConfig.MaxUserKeys - 1}
	svc := NewApiKeyService(repo, repo, DefaultApiKeyConfig)
	got, credential, err := svc.CreateUserKey(context.Background(), 8, " local ")
	if err != nil {
		t.Fatalf("CreateUserKey: %v", err)
	}
	if got.ID <= 0 || got.ID != credential.KeyID {
		t.Fatalf("ID = %d, credential key id = %d", got.ID, credential.KeyID)
	}
	if repo.Created == nil || repo.Created.UserID != 8 || repo.Created.Role != ApiKeyRoleUser {
		t.Fatalf("created key = %+v", repo.Created)
	}
	if repo.Created.SecretHash == "" || repo.Created.SecretHash != credential.SecretHash {
		t.Fatalf("created secret hash = %q, want credential hash", repo.Created.SecretHash)
	}
	if got.Name != "local" {
		t.Fatalf("Name = %q, want %q", got.Name, "local")
	}
}

func TestApiKeyService_CreateUserKeyRejectsLimit(t *testing.T) {
	repo := &MockApiKeyRepository{Count: DefaultApiKeyConfig.MaxUserKeys}
	svc := NewApiKeyService(repo, repo, DefaultApiKeyConfig)
	_, _, err := svc.CreateUserKey(context.Background(), 8, " local ")
	if !errors.Is(err, ErrApiKeyLimitExceeded) {
		t.Fatalf("CreateUserKey error = %v, want ErrApiKeyLimitExceeded", err)
	}
	if repo.Created != nil {
		t.Fatalf("created key = %+v, want nil", repo.Created)
	}
}

func TestApiKeyService_DeleteUserKey(t *testing.T) {
	repo := &MockApiKeyRepository{Key: &ApiKey{ID: 11, Role: ApiKeyRoleUser, UserID: 8}, DeleteOK: true}
	svc := NewApiKeyService(repo, repo, DefaultApiKeyConfig)
	if err := svc.DeleteUserKey(context.Background(), 8, 11); err != nil {
		t.Fatalf("DeleteUserKey: %v", err)
	}
	if repo.DeleteID != 11 {
		t.Fatalf("delete id = %d, want 11", repo.DeleteID)
	}
}

func TestApiKeyService_DeleteUserKeyMissing(t *testing.T) {
	repo := &MockApiKeyRepository{Key: &ApiKey{ID: 11, Role: ApiKeyRoleUser, UserID: 8}, DeleteOK: false}
	svc := NewApiKeyService(repo, repo, DefaultApiKeyConfig)
	err := svc.DeleteUserKey(context.Background(), 8, 11)
	if !errors.Is(err, ErrApiKeyNotFound) {
		t.Fatalf("DeleteUserKey error = %v, want ErrApiKeyNotFound", err)
	}
}

func testApiKeyCredential(keyID int64) ApiKeyCredential {
	credential := ApiKeyCredential{KeyID: keyID, Secret: []byte("1234567890abcdef")}
	credential.SecretHash = credential.HashSecret()
	return credential
}
