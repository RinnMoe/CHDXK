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

type apiKeyServiceFakeRepo struct {
	key       *ApiKey
	keys      []ApiKey
	count     int
	created   *ApiKey
	deleted   bool
	touched   bool
	deleteOK  bool
	deleteID  int
	deleteUID int
	findErr   error
	countErr  error
	createErr error
	deleteErr error
	touchErr  error
}

func (r *apiKeyServiceFakeRepo) FindByKey(_ context.Context, key string) (*ApiKey, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.key != nil && r.key.Key == key {
		copy := *r.key
		return &copy, nil
	}
	return nil, nil
}

func (r *apiKeyServiceFakeRepo) ListByUser(_ context.Context, _ int) ([]ApiKey, error) {
	return r.keys, nil
}

func (r *apiKeyServiceFakeRepo) CountByUser(_ context.Context, _ int) (int, error) {
	if r.countErr != nil {
		return 0, r.countErr
	}
	return r.count, nil
}

func (r *apiKeyServiceFakeRepo) Create(_ context.Context, apiKey *ApiKey) error {
	if r.createErr != nil {
		return r.createErr
	}
	copy := *apiKey
	r.created = &copy
	apiKey.ID = 99
	return nil
}

func (r *apiKeyServiceFakeRepo) DeleteByUser(_ context.Context, id int, userID int) (bool, error) {
	r.deleteID = id
	r.deleteUID = userID
	if r.deleteErr != nil {
		return false, r.deleteErr
	}
	return r.deleteOK, nil
}

func (r *apiKeyServiceFakeRepo) TouchLastUsed(_ context.Context, _ int, _ time.Time) error {
	r.touched = true
	return r.touchErr
}

func TestApiKeyService_ValidateKeyReturnsKey(t *testing.T) {
	repo := &apiKeyServiceFakeRepo{key: &ApiKey{ID: 1, Key: "k1", Role: ApiKeyRoleSystem}}
	svc := NewApiKeyService(repo)
	got, err := svc.ValidateKey(context.Background(), "k1")
	if err != nil {
		t.Fatalf("ValidateKey: %v", err)
	}
	if got == nil || got.ID != 1 {
		t.Fatalf("ValidateKey = %+v, want id 1", got)
	}
}

func TestApiKeyService_MarkKeyUsedTouchesLastUsed(t *testing.T) {
	repo := &apiKeyServiceFakeRepo{}
	svc := NewApiKeyService(repo)
	if err := svc.MarkKeyUsed(context.Background(), 1); err != nil {
		t.Fatalf("MarkKeyUsed: %v", err)
	}
	if !repo.touched {
		t.Fatal("expected TouchLastUsed to be called")
	}
}

func TestApiKeyService_CreateUserKey(t *testing.T) {
	repo := &apiKeyServiceFakeRepo{count: MaxUserApiKeyCount - 1}
	svc := NewApiKeyService(repo)
	got, err := svc.CreateUserKey(context.Background(), 8, " local ")
	if err != nil {
		t.Fatalf("CreateUserKey: %v", err)
	}
	if got.ID != 99 {
		t.Fatalf("ID = %d, want 99", got.ID)
	}
	if repo.created == nil || repo.created.UserID != 8 || repo.created.Role != ApiKeyRoleUser {
		t.Fatalf("created key = %+v", repo.created)
	}
	if got.Name != "local" {
		t.Fatalf("Name = %q, want %q", got.Name, "local")
	}
}

func TestApiKeyService_CreateUserKeyRejectsLimit(t *testing.T) {
	repo := &apiKeyServiceFakeRepo{count: MaxUserApiKeyCount}
	svc := NewApiKeyService(repo)
	_, err := svc.CreateUserKey(context.Background(), 8, " local ")
	if !errors.Is(err, ErrApiKeyLimitExceeded) {
		t.Fatalf("CreateUserKey error = %v, want ErrApiKeyLimitExceeded", err)
	}
	if repo.created != nil {
		t.Fatalf("created key = %+v, want nil", repo.created)
	}
}

func TestApiKeyService_DeleteUserKey(t *testing.T) {
	repo := &apiKeyServiceFakeRepo{deleteOK: true}
	svc := NewApiKeyService(repo)
	if err := svc.DeleteUserKey(context.Background(), 8, 11); err != nil {
		t.Fatalf("DeleteUserKey: %v", err)
	}
	if repo.deleteID != 11 || repo.deleteUID != 8 {
		t.Fatalf("delete args = (%d,%d), want (11,8)", repo.deleteID, repo.deleteUID)
	}
}

func TestApiKeyService_DeleteUserKeyMissing(t *testing.T) {
	repo := &apiKeyServiceFakeRepo{deleteOK: false}
	svc := NewApiKeyService(repo)
	err := svc.DeleteUserKey(context.Background(), 8, 11)
	if !errors.Is(err, ErrApiKeyNotFound) {
		t.Fatalf("DeleteUserKey error = %v, want ErrApiKeyNotFound", err)
	}
}
