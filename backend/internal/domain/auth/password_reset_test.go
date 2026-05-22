package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPasswordResetService_SendResetCodeSuccess(t *testing.T) {
	repo := newResetFakeUserRepo(map[string]*User{
		"alice@example.edu": {ID: 1, Email: "alice@example.edu"},
	})
	codes := newResetFakeCodeRepo()
	sender := &resetFakeSender{}
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	svc := NewPasswordResetService(repo, codes, sender, hasher, PasswordResetConfig{
		CodeInterval: time.Minute,
		CodeTTL:      10 * time.Minute,
	})

	if err := svc.SendResetCode(context.Background(), "alice@example.edu"); err != nil {
		t.Fatalf("SendResetCode: %v", err)
	}
	if sender.email != "alice@example.edu" {
		t.Fatalf("sender email = %q, want alice@example.edu", sender.email)
	}
	if len(sender.code) != 6 {
		t.Fatalf("sender code = %q, want 6 digits", sender.code)
	}
}

func TestPasswordResetService_SendResetCodeRejectsUnknownUser(t *testing.T) {
	repo := newResetFakeUserRepo(nil)
	svc := NewPasswordResetService(repo, newResetFakeCodeRepo(), &resetFakeSender{}, nil, PasswordResetConfig{})

	err := svc.SendResetCode(context.Background(), "nobody@example.edu")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("SendResetCode error = %v, want ErrUserNotFound", err)
	}
}

func TestPasswordResetService_SendResetCodeRateLimit(t *testing.T) {
	repo := newResetFakeUserRepo(map[string]*User{
		"alice@example.edu": {ID: 1, Email: "alice@example.edu"},
	})
	svc := NewPasswordResetService(repo, newResetFakeCodeRepo(), &resetFakeSender{}, nil, PasswordResetConfig{
		CodeInterval: time.Minute,
		CodeTTL:      10 * time.Minute,
	})
	ctx := context.Background()

	if err := svc.SendResetCode(ctx, "alice@example.edu"); err != nil {
		t.Fatalf("first SendResetCode: %v", err)
	}
	err := svc.SendResetCode(ctx, "alice@example.edu")
	if !errors.Is(err, ErrVerificationTooSoon) {
		t.Fatalf("second SendResetCode error = %v, want ErrVerificationTooSoon", err)
	}
}

func TestPasswordResetService_ResetPasswordSuccess(t *testing.T) {
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(1)
	oldHash, err := hasher.Hash("oldpass")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	repo := newResetFakeUserRepo(map[string]*User{
		"alice@example.edu": {ID: 1, Email: "alice@example.edu", Password: oldHash},
	})
	codes := newResetFakeCodeRepo()
	codes.saved["alice@example.edu"] = VerificationCode{
		Email: "alice@example.edu", Code: "123456", ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	svc := NewPasswordResetService(repo, codes, &resetFakeSender{}, hasher, PasswordResetConfig{})

	if err := svc.ResetPassword(context.Background(), "alice@example.edu", "123456", "newpass"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	if !hasher.Verify("newpass", repo.users["alice@example.edu"].Password) {
		t.Fatal("password was not updated")
	}
	if _, exists := codes.saved["alice@example.edu"]; exists {
		t.Fatal("verification code should be deleted after reset")
	}
}

func TestPasswordResetService_ResetPasswordRejectsInvalidCode(t *testing.T) {
	repo := newResetFakeUserRepo(map[string]*User{
		"alice@example.edu": {ID: 1, Email: "alice@example.edu"},
	})
	codes := newResetFakeCodeRepo()
	codes.saved["alice@example.edu"] = VerificationCode{
		Email: "alice@example.edu", Code: "123456", ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	svc := NewPasswordResetService(repo, codes, &resetFakeSender{}, NewDjangoPBKDF2SHA256PasswordHasher(1), PasswordResetConfig{})

	err := svc.ResetPassword(context.Background(), "alice@example.edu", "000000", "newpass")
	if !errors.Is(err, ErrVerificationCodeInvalid) {
		t.Fatalf("ResetPassword error = %v, want ErrVerificationCodeInvalid", err)
	}
}

func TestPasswordResetService_ResetPasswordRejectsEmptyPassword(t *testing.T) {
	repo := newResetFakeUserRepo(map[string]*User{
		"alice@example.edu": {ID: 1, Email: "alice@example.edu"},
	})
	svc := NewPasswordResetService(repo, newResetFakeCodeRepo(), &resetFakeSender{}, nil, PasswordResetConfig{})

	err := svc.ResetPassword(context.Background(), "alice@example.edu", "123456", "  ")
	if !errors.Is(err, ErrPasswordRequired) {
		t.Fatalf("ResetPassword error = %v, want ErrPasswordRequired", err)
	}
}

func TestPasswordResetService_ResetPasswordRejectsUnknownUser(t *testing.T) {
	codes := newResetFakeCodeRepo()
	codes.saved["nobody@example.edu"] = VerificationCode{
		Email: "nobody@example.edu", Code: "123456", ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	svc := NewPasswordResetService(newResetFakeUserRepo(nil), codes, &resetFakeSender{}, NewDjangoPBKDF2SHA256PasswordHasher(1), PasswordResetConfig{})

	err := svc.ResetPassword(context.Background(), "nobody@example.edu", "123456", "newpass")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("ResetPassword error = %v, want ErrUserNotFound", err)
	}
}

// --- fakes for password_reset tests ---

type resetFakeUserRepo struct {
	users map[string]*User
}

func newResetFakeUserRepo(users map[string]*User) *resetFakeUserRepo {
	if users == nil {
		users = map[string]*User{}
	}
	return &resetFakeUserRepo{users: users}
}

func (r *resetFakeUserRepo) Create(_ context.Context, u *User) error {
	r.users[u.Email] = u
	return nil
}

func (r *resetFakeUserRepo) Update(_ context.Context, u *User) error {
	r.users[u.Email] = u
	return nil
}

func (r *resetFakeUserRepo) TouchLastSeen(_ context.Context, _ int, _ time.Time) error { return nil }

func (r *resetFakeUserRepo) FindByID(_ context.Context, _ int) (*User, error) { return nil, nil }

func (r *resetFakeUserRepo) FindByUsername(_ context.Context, _ string) (*User, error) {
	return nil, nil
}

func (r *resetFakeUserRepo) FindByEmail(_ context.Context, email string) (*User, error) {
	u, ok := r.users[email]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

type resetFakeCodeRepo struct {
	saved        map[string]VerificationCode
	cooldownTill map[string]time.Time
}

func newResetFakeCodeRepo() *resetFakeCodeRepo {
	return &resetFakeCodeRepo{saved: map[string]VerificationCode{}, cooldownTill: map[string]time.Time{}}
}

func (r *resetFakeCodeRepo) ReserveSend(_ context.Context, email string, interval time.Duration) (time.Duration, error) {
	now := time.Now()
	if till := r.cooldownTill[email]; till.After(now) {
		return time.Until(till), nil
	}
	r.cooldownTill[email] = now.Add(interval)
	return 0, nil
}

func (r *resetFakeCodeRepo) Save(_ context.Context, code VerificationCode, _ time.Duration) error {
	r.saved[code.Email] = code
	return nil
}

func (r *resetFakeCodeRepo) Get(_ context.Context, email string) (*VerificationCode, error) {
	c, ok := r.saved[email]
	if !ok {
		return nil, nil
	}
	return &c, nil
}

func (r *resetFakeCodeRepo) Delete(_ context.Context, email string) error {
	delete(r.saved, email)
	return nil
}

type resetFakeSender struct {
	email string
	code  string
}

func (s *resetFakeSender) SendVerificationCode(_ context.Context, email string, code string) error {
	s.email = email
	s.code = code
	return nil
}
