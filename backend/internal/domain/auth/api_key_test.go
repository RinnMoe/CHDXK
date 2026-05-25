package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGenerateApiKey(t *testing.T) {
	key, err := GenerateApiKey()
	if err != nil {
		t.Fatalf("GenerateApiKey: %v", err)
	}
	if len(key) != len(apiKeyPrefix)+64 {
		t.Fatalf("len(key) = %d, want %d", len(key), len(apiKeyPrefix)+64)
	}
	if key[:len(apiKeyPrefix)] != apiKeyPrefix {
		t.Fatalf("prefix = %q, want %q", key[:len(apiKeyPrefix)], apiKeyPrefix)
	}
}

func TestNewUserApiKey(t *testing.T) {
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
	key, err := NewUserApiKey("  local script  ", 12, now)
	if err != nil {
		t.Fatalf("NewUserApiKey: %v", err)
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
	if key.Key == "" {
		t.Fatal("expected generated key")
	}
}

func TestNewUserApiKeyRejectsEmptyName(t *testing.T) {
	_, err := NewUserApiKey("  ", 12, time.Now())
	if !errors.Is(err, ErrApiKeyNameRequired) {
		t.Fatalf("err = %v, want ErrApiKeyNameRequired", err)
	}
}

func TestMaskApiKey(t *testing.T) {
	if got, want := MaskApiKey("short"), "*****"; got != want {
		t.Fatalf("MaskApiKey(short) = %q, want %q", got, want)
	}
	if got, want := MaskApiKey("jc_1234567890abcdef"), "jc_1234************abcdef"; got != want {
		t.Fatalf("MaskApiKey(long) = %q, want %q", got, want)
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
	repo := &MockApiKeyRepository{Key: &ApiKey{ID: 1, Key: "k1", Role: ApiKeyRoleSystem}}
	svc := NewApiKeyService(repo, DefaultApiKeyConfig)
	got, err := svc.ValidateKey(context.Background(), "k1")
	if err != nil {
		t.Fatalf("ValidateKey: %v", err)
	}
	if got == nil || got.ID != 1 {
		t.Fatalf("ValidateKey = %+v, want id 1", got)
	}
}

func TestApiKeyService_CreateUserKey(t *testing.T) {
	repo := &MockApiKeyRepository{Count: DefaultApiKeyConfig.MaxUserKeys - 1}
	svc := NewApiKeyService(repo, DefaultApiKeyConfig)
	got, err := svc.CreateUserKey(context.Background(), 8, " local ")
	if err != nil {
		t.Fatalf("CreateUserKey: %v", err)
	}
	if got.ID != 99 {
		t.Fatalf("ID = %d, want 99", got.ID)
	}
	if repo.Created == nil || repo.Created.UserID != 8 || repo.Created.Role != ApiKeyRoleUser {
		t.Fatalf("created key = %+v", repo.Created)
	}
	if got.Name != "local" {
		t.Fatalf("Name = %q, want %q", got.Name, "local")
	}
}

func TestApiKeyService_CreateUserKeyRejectsLimit(t *testing.T) {
	repo := &MockApiKeyRepository{Count: DefaultApiKeyConfig.MaxUserKeys}
	svc := NewApiKeyService(repo, DefaultApiKeyConfig)
	_, err := svc.CreateUserKey(context.Background(), 8, " local ")
	if !errors.Is(err, ErrApiKeyLimitExceeded) {
		t.Fatalf("CreateUserKey error = %v, want ErrApiKeyLimitExceeded", err)
	}
	if repo.Created != nil {
		t.Fatalf("created key = %+v, want nil", repo.Created)
	}
}

func TestApiKeyService_DeleteUserKey(t *testing.T) {
	repo := &MockApiKeyRepository{Key: &ApiKey{ID: 11, Role: ApiKeyRoleUser, UserID: 8}, DeleteOK: true}
	svc := NewApiKeyService(repo, DefaultApiKeyConfig)
	if err := svc.DeleteUserKey(context.Background(), 8, 11); err != nil {
		t.Fatalf("DeleteUserKey: %v", err)
	}
	if repo.DeleteID != 11 {
		t.Fatalf("delete id = %d, want 11", repo.DeleteID)
	}
}

func TestApiKeyService_DeleteUserKeyMissing(t *testing.T) {
	repo := &MockApiKeyRepository{Key: &ApiKey{ID: 11, Role: ApiKeyRoleUser, UserID: 8}, DeleteOK: false}
	svc := NewApiKeyService(repo, DefaultApiKeyConfig)
	err := svc.DeleteUserKey(context.Background(), 8, 11)
	if !errors.Is(err, ErrApiKeyNotFound) {
		t.Fatalf("DeleteUserKey error = %v, want ErrApiKeyNotFound", err)
	}
}
