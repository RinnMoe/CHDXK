package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/account/security"
	"jcourse/pkg/apperr"
)

func TestLoginService_LoginLockedAfterMaxAttempts(t *testing.T) {
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{username: {ID: 1, Username: username, PasswordHash: password}})
	attempts := security.NewMockLoginAttemptRepository(map[string]int{"alice@example.edu": 5})
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), LoginConfig{MaxAttempts: 5, Lockout: 15 * time.Minute})

	_, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if !errors.Is(err, security.ErrLoginLocked) {
		t.Fatalf("Login error = %v, want ErrLoginLocked", err)
	}
}

func TestLoginService_LoginNotLockedWhenBelowMax(t *testing.T) {
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{username: {ID: 1, Username: username, PasswordHash: password}})
	attempts := security.NewMockLoginAttemptRepository(map[string]int{"alice@example.edu": 4})
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), LoginConfig{MaxAttempts: 5, Lockout: 15 * time.Minute})

	u, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if u.ID != 1 {
		t.Fatalf("Login ID = %d, want 1", u.ID)
	}
}

func TestLoginService_FailedLoginIncrementsAttempts(t *testing.T) {
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{username: {ID: 1, Username: username, PasswordHash: password}})
	attempts := security.NewMockLoginAttemptRepository(map[string]int{})
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), LoginConfig{MaxAttempts: 5, Lockout: 15 * time.Minute})

	_, err := svc.Login(context.Background(), "alice@example.edu", "wrong")
	if !errors.Is(err, security.ErrInvalidCredentials) {
		t.Fatalf("Login error = %v, want ErrInvalidCredentials", err)
	}
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Login error = %T, want AppError", err)
	}
	if appErr.Msg != "邮箱或密码错误，还有 4 次尝试机会" {
		t.Fatalf("Login error message = %q, want remaining attempts", appErr.Msg)
	}
	if attempts.Counts["alice@example.edu"] != 1 {
		t.Fatalf("attempt count = %d, want 1", attempts.Counts["alice@example.edu"])
	}
}

func TestLoginService_LoginRejectsAccountWithoutPassword(t *testing.T) {
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{username: {ID: 1, Username: username, PasswordHash: ""}})
	attempts := security.NewMockLoginAttemptRepository(map[string]int{})
	svc := NewLoginService(repo, credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1}), attempts, testUsernameDeriver(), LoginConfig{MaxAttempts: 5, Lockout: 15 * time.Minute})

	_, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if !errors.Is(err, security.ErrPasswordNotSet) {
		t.Fatalf("Login error = %v, want ErrPasswordNotSet", err)
	}
	if attempts.Counts["alice@example.edu"] != 0 {
		t.Fatalf("attempt count = %d, want 0", attempts.Counts["alice@example.edu"])
	}
}

func TestLoginService_SuccessfulLoginResetsAttempts(t *testing.T) {
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{username: {ID: 1, Username: username, PasswordHash: password}})
	attempts := security.NewMockLoginAttemptRepository(map[string]int{"alice@example.edu": 3})
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), LoginConfig{MaxAttempts: 5, Lockout: 15 * time.Minute})

	_, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if attempts.Counts["alice@example.edu"] != 0 {
		t.Fatalf("attempt count after success = %d, want 0", attempts.Counts["alice@example.edu"])
	}
}

func TestLoginService_DefaultConfigAppliesLockout(t *testing.T) {
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{username: {ID: 1, Username: username, PasswordHash: password}})
	attempts := security.NewMockLoginAttemptRepository(map[string]int{"alice@example.edu": 999})
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), LoginConfig{})

	_, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if !errors.Is(err, security.ErrLoginLocked) {
		t.Fatalf("Login error = %v, want ErrLoginLocked", err)
	}
}

func TestLoginService_LockoutTriggersOnNthFailure(t *testing.T) {
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{username: {ID: 1, Username: username, PasswordHash: "irrelevant"}})
	attempts := security.NewMockLoginAttemptRepository(map[string]int{})
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), LoginConfig{MaxAttempts: 3, Lockout: 15 * time.Minute})
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		_, err := svc.Login(ctx, "alice@example.edu", "wrong")
		if !errors.Is(err, security.ErrInvalidCredentials) {
			t.Fatalf("attempt %d error = %v, want ErrInvalidCredentials", i, err)
		}
	}

	_, err := svc.Login(ctx, "alice@example.edu", "wrong")
	if !errors.Is(err, security.ErrLoginLocked) {
		t.Fatalf("error after 3 failures = %v, want ErrLoginLocked", err)
	}
}

func mustUsernameFromEmail(t *testing.T, email string) string {
	t.Helper()
	username, err := testUsernameDeriver().UsernameFromEmail(email)
	if err != nil {
		t.Fatalf("UsernameFromEmail: %v", err)
	}
	return username
}

func testUsernameDeriver() identity.UsernameDeriver {
	return identity.NewBLAKE2bUsernameDeriver(identity.UsernameDeriverConfig{Salt: "SALT"})
}
