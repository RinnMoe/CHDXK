package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAuthenticationService_LoginLockedAfterMaxAttempts(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, _ := hasher.Hash("secret")
	repo := &lockFakeUserRepo{
		users: map[string]*User{
			"alice@example.edu": {ID: 1, Email: "alice@example.edu", Password: password, Role: RoleUser},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{"alice@example.edu": 5}}
	svc := NewAuthenticationService(repo, hasher, attempts, 5, 15*time.Minute)

	_, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if !errors.Is(err, ErrLoginLocked) {
		t.Fatalf("Login error = %v, want ErrLoginLocked", err)
	}
}

func TestAuthenticationService_LoginNotLockedWhenBelowMax(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, _ := hasher.Hash("secret")
	repo := &lockFakeUserRepo{
		users: map[string]*User{
			"alice@example.edu": {ID: 1, Email: "alice@example.edu", Password: password, Role: RoleUser},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{"alice@example.edu": 4}}
	svc := NewAuthenticationService(repo, hasher, attempts, 5, 15*time.Minute)

	u, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if u.ID != 1 {
		t.Fatalf("Login ID = %d, want 1", u.ID)
	}
}

func TestAuthenticationService_FailedLoginIncrementsAttempts(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, _ := hasher.Hash("secret")
	repo := &lockFakeUserRepo{
		users: map[string]*User{
			"alice@example.edu": {ID: 1, Email: "alice@example.edu", Password: password, Role: RoleUser},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{}}
	svc := NewAuthenticationService(repo, hasher, attempts, 5, 15*time.Minute)

	_, err := svc.Login(context.Background(), "alice@example.edu", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login error = %v, want ErrInvalidCredentials", err)
	}
	if attempts.counts["alice@example.edu"] != 1 {
		t.Fatalf("attempt count = %d, want 1", attempts.counts["alice@example.edu"])
	}
}

func TestAuthenticationService_SuccessfulLoginResetsAttempts(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, _ := hasher.Hash("secret")
	repo := &lockFakeUserRepo{
		users: map[string]*User{
			"alice@example.edu": {ID: 1, Email: "alice@example.edu", Password: password, Role: RoleUser},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{"alice@example.edu": 3}}
	svc := NewAuthenticationService(repo, hasher, attempts, 5, 15*time.Minute)

	_, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if attempts.counts["alice@example.edu"] != 0 {
		t.Fatalf("attempt count after success = %d, want 0", attempts.counts["alice@example.edu"])
	}
}

func TestAuthenticationService_NoLockoutWhenMaxAttemptsZero(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, _ := hasher.Hash("secret")
	repo := &lockFakeUserRepo{
		users: map[string]*User{
			"alice@example.edu": {ID: 1, Email: "alice@example.edu", Password: password, Role: RoleUser},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{"alice@example.edu": 999}}
	svc := NewAuthenticationService(repo, hasher, attempts, 0, 15*time.Minute)

	u, err := svc.Login(context.Background(), "alice@example.edu", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if u.ID != 1 {
		t.Fatalf("Login ID = %d, want 1", u.ID)
	}
}

func TestAuthenticationService_LockoutTriggersOnNthFailure(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	repo := &lockFakeUserRepo{
		users: map[string]*User{
			"alice@example.edu": {ID: 1, Email: "alice@example.edu", Password: "irrelevant", Role: RoleUser},
		},
	}
	attempts := &lockFakeAttempts{counts: map[string]int{}}
	svc := NewAuthenticationService(repo, hasher, attempts, 3, 15*time.Minute)
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

// --- fakes for authentication lockout tests ---

type lockFakeUserRepo struct {
	users map[string]*User
}

func (r *lockFakeUserRepo) Create(_ context.Context, u *User) error {
	r.users[u.Email] = u
	return nil
}

func (r *lockFakeUserRepo) Update(_ context.Context, u *User) error {
	r.users[u.Email] = u
	return nil
}

func (r *lockFakeUserRepo) TouchLastSeen(_ context.Context, _ int, _ time.Time) error { return nil }

func (r *lockFakeUserRepo) FindByID(_ context.Context, _ int) (*User, error) { return nil, nil }

func (r *lockFakeUserRepo) FindByUsername(_ context.Context, _ string) (*User, error) { return nil, nil }

func (r *lockFakeUserRepo) FindByEmail(_ context.Context, email string) (*User, error) {
	u, ok := r.users[email]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
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
