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

func TestNewSystemApiKeyRejectsEmptyName(t *testing.T) {
	_, err := NewSystemApiKey("  ", testApiKeyCredential(12), time.Now())
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
	if !(ApiKey{Role: ApiKeyRoleSystem}).IsSystem() {
		t.Fatal("expected system key to be system")
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

func testApiKeyCredential(keyID int64) ApiKeyCredential {
	credential := ApiKeyCredential{KeyID: keyID, Secret: []byte("1234567890abcdef")}
	credential.SecretHash = credential.HashSecret()
	return credential
}
