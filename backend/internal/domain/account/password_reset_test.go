package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/account/notification"
	"jcourse/internal/domain/account/verification"
	domainemail "jcourse/internal/domain/email"
)

func TestPasswordResetService_SendResetCodeSuccess(t *testing.T) {
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{
		username: {ID: 1, Username: username},
	})
	codes := verification.NewMockCodeRepository()
	sender := &resetFakeSender{}
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	svc := NewPasswordResetService(repo, codes, sender, hasher, testUsernameDeriver(), verification.Config{
		CodeInterval: time.Minute,
		CodeTTL:      10 * time.Minute,
	})

	if err := svc.SendResetCode(context.Background(), "alice@example.edu"); err != nil {
		t.Fatalf("SendResetCode: %v", err)
	}
	if sender.email.To != "alice@example.edu" {
		t.Fatalf("sender email.To = %q, want alice@example.edu", sender.email.To)
	}
	if sender.email.Subject != notification.VerificationCodeEmailSubject {
		t.Fatalf("sender email.Subject = %q, want %q", sender.email.Subject, notification.VerificationCodeEmailSubject)
	}
	saved := codes.Saved["alice@example.edu"]
	if len(saved.Code) != 6 {
		t.Fatalf("saved code = %q, want 6 digits", saved.Code)
	}
}

func TestPasswordResetService_SendResetCodeRejectsUnknownUser(t *testing.T) {
	repo := identity.NewMockRepository(nil)
	svc := NewPasswordResetService(repo, verification.NewMockCodeRepository(), &resetFakeSender{}, nil, testUsernameDeriver(), verification.Config{})

	err := svc.SendResetCode(context.Background(), "nobody@example.edu")
	if !errors.Is(err, identity.ErrNotFound) {
		t.Fatalf("SendResetCode error = %v, want ErrUserNotFound", err)
	}
}

func TestPasswordResetService_SendResetCodeRateLimit(t *testing.T) {
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{
		username: {ID: 1, Username: username},
	})
	svc := NewPasswordResetService(repo, verification.NewMockCodeRepository(), &resetFakeSender{}, nil, testUsernameDeriver(), verification.Config{
		CodeInterval: time.Minute,
		CodeTTL:      10 * time.Minute,
	})
	ctx := context.Background()

	if err := svc.SendResetCode(ctx, "alice@example.edu"); err != nil {
		t.Fatalf("first SendResetCode: %v", err)
	}
	err := svc.SendResetCode(ctx, "alice@example.edu")
	if !errors.Is(err, verification.ErrSendTooSoon) {
		t.Fatalf("second SendResetCode error = %v, want ErrVerificationTooSoon", err)
	}
}

func TestPasswordResetService_ResetPasswordSuccess(t *testing.T) {
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	oldHash, err := hasher.Hash("oldpass")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{
		username: {ID: 1, Username: username, PasswordHash: oldHash},
	})
	codes := verification.NewMockCodeRepository()
	codes.Saved["alice@example.edu"] = verification.Code{
		Email: "alice@example.edu", Code: "123456", ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	svc := NewPasswordResetService(repo, codes, &resetFakeSender{}, hasher, testUsernameDeriver(), verification.Config{})

	if err := svc.ResetPassword(context.Background(), "alice@example.edu", "123456", "newpass"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	if !hasher.Verify("newpass", repo.AccountsByUsername[username].PasswordHash) {
		t.Fatal("password was not updated")
	}
	if _, exists := codes.Saved["alice@example.edu"]; exists {
		t.Fatal("verification code should be deleted after reset")
	}
}

func TestPasswordResetService_ResetPasswordRejectsInvalidCode(t *testing.T) {
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{
		username: {ID: 1, Username: username},
	})
	codes := verification.NewMockCodeRepository()
	codes.Saved["alice@example.edu"] = verification.Code{
		Email: "alice@example.edu", Code: "123456", ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	svc := NewPasswordResetService(repo, codes, &resetFakeSender{}, credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1}), testUsernameDeriver(), verification.Config{})

	err := svc.ResetPassword(context.Background(), "alice@example.edu", "000000", "newpass")
	if !errors.Is(err, verification.ErrCodeInvalid) {
		t.Fatalf("ResetPassword error = %v, want ErrVerificationCodeInvalid", err)
	}
}

func TestPasswordResetService_ResetPasswordRejectsEmptyPassword(t *testing.T) {
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{
		username: {ID: 1, Username: username},
	})
	svc := NewPasswordResetService(repo, verification.NewMockCodeRepository(), &resetFakeSender{}, nil, testUsernameDeriver(), verification.Config{})

	err := svc.ResetPassword(context.Background(), "alice@example.edu", "123456", "  ")
	if !errors.Is(err, credential.ErrPasswordRequired) {
		t.Fatalf("ResetPassword error = %v, want ErrPasswordRequired", err)
	}
}

func TestPasswordResetService_ResetPasswordRejectsUnknownUser(t *testing.T) {
	codes := verification.NewMockCodeRepository()
	codes.Saved["nobody@example.edu"] = verification.Code{
		Email: "nobody@example.edu", Code: "123456", ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	svc := NewPasswordResetService(identity.NewMockRepository(nil), codes, &resetFakeSender{}, credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1}), testUsernameDeriver(), verification.Config{})

	err := svc.ResetPassword(context.Background(), "nobody@example.edu", "123456", "newpass")
	if !errors.Is(err, identity.ErrNotFound) {
		t.Fatalf("ResetPassword error = %v, want ErrUserNotFound", err)
	}
}

type resetFakeSender struct {
	email domainemail.Email
}

func (s *resetFakeSender) SendEmail(_ context.Context, email domainemail.Email) error {
	s.email = email
	return nil
}
