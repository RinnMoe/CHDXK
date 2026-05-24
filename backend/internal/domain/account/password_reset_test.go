package account

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPasswordResetService_SendResetCodeSuccess(t *testing.T) {
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := newResetFakeUserRepo(map[string]*Account{
		username: {ID: 1, Username: username},
	})
	codes := newResetFakeCodeRepo()
	sender := &resetFakeSender{}
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(PasswordHashConfig{Iterations: 1})
	svc := NewPasswordResetService(repo, codes, sender, hasher, testUsernameDeriver(), PasswordResetConfig{
		CodeInterval: time.Minute,
		CodeTTL:      10 * time.Minute,
	})

	if err := svc.SendResetCode(context.Background(), "alice@example.edu"); err != nil {
		t.Fatalf("SendResetCode: %v", err)
	}
	if sender.email.To != "alice@example.edu" {
		t.Fatalf("sender email.To = %q, want alice@example.edu", sender.email.To)
	}
	if sender.email.Subject != VerificationCodeEmailSubject {
		t.Fatalf("sender email.Subject = %q, want %q", sender.email.Subject, VerificationCodeEmailSubject)
	}
	saved := codes.saved["alice@example.edu"]
	if len(saved.Code) != 6 {
		t.Fatalf("saved code = %q, want 6 digits", saved.Code)
	}
}

func TestPasswordResetService_SendResetCodeRejectsUnknownUser(t *testing.T) {
	repo := newResetFakeUserRepo(nil)
	svc := NewPasswordResetService(repo, newResetFakeCodeRepo(), &resetFakeSender{}, nil, testUsernameDeriver(), PasswordResetConfig{})

	err := svc.SendResetCode(context.Background(), "nobody@example.edu")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("SendResetCode error = %v, want ErrUserNotFound", err)
	}
}

func TestPasswordResetService_SendResetCodeRateLimit(t *testing.T) {
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := newResetFakeUserRepo(map[string]*Account{
		username: {ID: 1, Username: username},
	})
	svc := NewPasswordResetService(repo, newResetFakeCodeRepo(), &resetFakeSender{}, nil, testUsernameDeriver(), PasswordResetConfig{
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
	hasher := NewDjangoPBKDF2SHA256PasswordHasher(PasswordHashConfig{Iterations: 1})
	oldHash, err := hasher.Hash("oldpass")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := newResetFakeUserRepo(map[string]*Account{
		username: {ID: 1, Username: username, PasswordHash: oldHash},
	})
	codes := newResetFakeCodeRepo()
	codes.saved["alice@example.edu"] = VerificationCode{
		Email: "alice@example.edu", Code: "123456", ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	svc := NewPasswordResetService(repo, codes, &resetFakeSender{}, hasher, testUsernameDeriver(), PasswordResetConfig{})

	if err := svc.ResetPassword(context.Background(), "alice@example.edu", "123456", "newpass"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	if !hasher.Verify("newpass", repo.users[username].PasswordHash) {
		t.Fatal("password was not updated")
	}
	if _, exists := codes.saved["alice@example.edu"]; exists {
		t.Fatal("verification code should be deleted after reset")
	}
}

func TestPasswordResetService_ResetPasswordRejectsInvalidCode(t *testing.T) {
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := newResetFakeUserRepo(map[string]*Account{
		username: {ID: 1, Username: username},
	})
	codes := newResetFakeCodeRepo()
	codes.saved["alice@example.edu"] = VerificationCode{
		Email: "alice@example.edu", Code: "123456", ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	svc := NewPasswordResetService(repo, codes, &resetFakeSender{}, NewDjangoPBKDF2SHA256PasswordHasher(PasswordHashConfig{Iterations: 1}), testUsernameDeriver(), PasswordResetConfig{})

	err := svc.ResetPassword(context.Background(), "alice@example.edu", "000000", "newpass")
	if !errors.Is(err, ErrVerificationCodeInvalid) {
		t.Fatalf("ResetPassword error = %v, want ErrVerificationCodeInvalid", err)
	}
}

func TestPasswordResetService_ResetPasswordRejectsEmptyPassword(t *testing.T) {
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := newResetFakeUserRepo(map[string]*Account{
		username: {ID: 1, Username: username},
	})
	svc := NewPasswordResetService(repo, newResetFakeCodeRepo(), &resetFakeSender{}, nil, testUsernameDeriver(), PasswordResetConfig{})

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
	svc := NewPasswordResetService(newResetFakeUserRepo(nil), codes, &resetFakeSender{}, NewDjangoPBKDF2SHA256PasswordHasher(PasswordHashConfig{Iterations: 1}), testUsernameDeriver(), PasswordResetConfig{})

	err := svc.ResetPassword(context.Background(), "nobody@example.edu", "123456", "newpass")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("ResetPassword error = %v, want ErrUserNotFound", err)
	}
}

// --- fakes for password_reset tests ---

type resetFakeUserRepo struct {
	users map[string]*Account
}

func newResetFakeUserRepo(users map[string]*Account) *resetFakeUserRepo {
	if users == nil {
		users = map[string]*Account{}
	}
	return &resetFakeUserRepo{users: users}
}

func (r *resetFakeUserRepo) Create(_ context.Context, u *Account) error {
	r.users[u.Username] = u
	return nil
}

func (r *resetFakeUserRepo) Update(_ context.Context, u *Account) error {
	r.users[u.Username] = u
	return nil
}

func (r *resetFakeUserRepo) TouchLastSeen(_ context.Context, _ int, _ time.Time) error { return nil }

func (r *resetFakeUserRepo) FindByID(_ context.Context, _ int) (*Account, error) { return nil, nil }

func (r *resetFakeUserRepo) FindByUsername(_ context.Context, username string) (*Account, error) {
	u, ok := r.users[username]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (r *resetFakeUserRepo) FindByEmail(_ context.Context, email string) (*Account, error) {
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
	email Email
}

func (s *resetFakeSender) SendEmail(_ context.Context, email Email) error {
	s.email = email
	return nil
}
