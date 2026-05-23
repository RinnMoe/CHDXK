package account

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLoginService_LoginLockedAfterMaxAttempts(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := &lockFakeUserRepo{
		users: map[string]*Account{
			username: {ID: 1, Username: username, PasswordHash: password},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{"alice@example.edu": 5}}
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), 5, 15*time.Minute)

	_, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if !errors.Is(err, ErrLoginLocked) {
		t.Fatalf("Login error = %v, want ErrLoginLocked", err)
	}
}

func TestLoginService_LoginNotLockedWhenBelowMax(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := &lockFakeUserRepo{
		users: map[string]*Account{
			username: {ID: 1, Username: username, PasswordHash: password},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{"alice@example.edu": 4}}
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), 5, 15*time.Minute)

	u, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if u.ID != 1 {
		t.Fatalf("Login ID = %d, want 1", u.ID)
	}
}

func TestLoginService_FailedLoginIncrementsAttempts(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := &lockFakeUserRepo{
		users: map[string]*Account{
			username: {ID: 1, Username: username, PasswordHash: password},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{}}
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), 5, 15*time.Minute)

	_, err := svc.Login(context.Background(), "alice@example.edu", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login error = %v, want ErrInvalidCredentials", err)
	}
	if attempts.counts["alice@example.edu"] != 1 {
		t.Fatalf("attempt count = %d, want 1", attempts.counts["alice@example.edu"])
	}
}

func TestLoginService_SuccessfulLoginResetsAttempts(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := &lockFakeUserRepo{
		users: map[string]*Account{
			username: {ID: 1, Username: username, PasswordHash: password},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{"alice@example.edu": 3}}
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), 5, 15*time.Minute)

	_, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if attempts.counts["alice@example.edu"] != 0 {
		t.Fatalf("attempt count after success = %d, want 0", attempts.counts["alice@example.edu"])
	}
}

func TestLoginService_NoLockoutWhenMaxAttemptsZero(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, _ := hasher.Hash("secret")
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := &lockFakeUserRepo{
		users: map[string]*Account{
			username: {ID: 1, Username: username, PasswordHash: password},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{"alice@example.edu": 999}}
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), 0, 15*time.Minute)

	u, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if u.ID != 1 {
		t.Fatalf("Login ID = %d, want 1", u.ID)
	}
}

func TestLoginService_LockoutTriggersOnNthFailure(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := &lockFakeUserRepo{
		users: map[string]*Account{
			username: {ID: 1, Username: username, PasswordHash: "irrelevant"},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{}}
	svc := NewLoginService(repo, hasher, attempts, testUsernameDeriver(), 3, 15*time.Minute)
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

// --- fakes for login lockout tests ---

type lockFakeUserRepo struct {
	users map[string]*Account
}

func (r *lockFakeUserRepo) Create(_ context.Context, u *Account) error {
	r.users[u.Username] = u
	return nil
}

func (r *lockFakeUserRepo) Update(_ context.Context, u *Account) error {
	r.users[u.Username] = u
	return nil
}

func (r *lockFakeUserRepo) TouchLastSeen(_ context.Context, _ int, _ time.Time) error { return nil }

func (r *lockFakeUserRepo) FindByID(_ context.Context, _ int) (*Account, error) { return nil, nil }

func (r *lockFakeUserRepo) FindByUsername(_ context.Context, username string) (*Account, error) {
	u, ok := r.users[username]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (r *lockFakeUserRepo) FindByEmail(_ context.Context, email string) (*Account, error) {
	u, ok := r.users[email]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
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
	return NewBLAKE2bUsernameDeriver("SALT")
}

type lockFakeAttempts struct {
	counts map[string]int
}

func (a *lockFakeAttempts) Increment(_ context.Context, email string) (int, error) {
	a.counts[email]++
	return a.counts[email], nil
}

func (a *lockFakeAttempts) Get(_ context.Context, email string) (int, error) {
	return a.counts[email], nil
}

func (a *lockFakeAttempts) Reset(_ context.Context, email string) error {
	a.counts[email] = 0
	return nil
}
