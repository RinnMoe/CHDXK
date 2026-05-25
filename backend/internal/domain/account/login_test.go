package account

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLoginService_LoginLockedAfterMaxAttempts(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(PasswordHashConfig{Iterations: 1})
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := NewMockAccountRepository(map[string]*Account{username: {ID: 1, Username: username, PasswordHash: password}})
	attempts := NewMockLoginAttemptRepository(map[string]int{"alice@example.edu": 5})
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), LoginConfig{MaxAttempts: 5, Lockout: 15 * time.Minute})

	_, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if !errors.Is(err, ErrLoginLocked) {
		t.Fatalf("Login error = %v, want ErrLoginLocked", err)
	}
}

func TestLoginService_LoginNotLockedWhenBelowMax(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(PasswordHashConfig{Iterations: 1})
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := NewMockAccountRepository(map[string]*Account{username: {ID: 1, Username: username, PasswordHash: password}})
	attempts := NewMockLoginAttemptRepository(map[string]int{"alice@example.edu": 4})
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
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(PasswordHashConfig{Iterations: 1})
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := NewMockAccountRepository(map[string]*Account{username: {ID: 1, Username: username, PasswordHash: password}})
	attempts := NewMockLoginAttemptRepository(map[string]int{})
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), LoginConfig{MaxAttempts: 5, Lockout: 15 * time.Minute})

	_, err := svc.Login(context.Background(), "alice@example.edu", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login error = %v, want ErrInvalidCredentials", err)
	}
	if attempts.Counts["alice@example.edu"] != 1 {
		t.Fatalf("attempt count = %d, want 1", attempts.Counts["alice@example.edu"])
	}
}

func TestLoginService_SuccessfulLoginResetsAttempts(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(PasswordHashConfig{Iterations: 1})
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := NewMockAccountRepository(map[string]*Account{username: {ID: 1, Username: username, PasswordHash: password}})
	attempts := NewMockLoginAttemptRepository(map[string]int{"alice@example.edu": 3})
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
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(PasswordHashConfig{Iterations: 1})
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := NewMockAccountRepository(map[string]*Account{username: {ID: 1, Username: username, PasswordHash: password}})
	attempts := NewMockLoginAttemptRepository(map[string]int{"alice@example.edu": 999})
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), LoginConfig{})

	_, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if !errors.Is(err, ErrLoginLocked) {
		t.Fatalf("Login error = %v, want ErrLoginLocked", err)
	}
}

func TestLoginService_LockoutTriggersOnNthFailure(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(PasswordHashConfig{Iterations: 1})
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := NewMockAccountRepository(map[string]*Account{username: {ID: 1, Username: username, PasswordHash: "irrelevant"}})
	attempts := NewMockLoginAttemptRepository(map[string]int{})
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), LoginConfig{MaxAttempts: 3, Lockout: 15 * time.Minute})
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		_, err := svc.Login(ctx, "alice@example.edu", "wrong")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("attempt %d error = %v, want ErrInvalidCredentials", i, err)
		}
	}

	_, err := svc.Login(ctx, "alice@example.edu", "wrong")
	if !errors.Is(err, ErrLoginLocked) {
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

func testUsernameDeriver() UsernameDeriver {
	return NewBLAKE2bUsernameDeriver(UsernameDeriverConfig{Salt: "SALT"})
}
