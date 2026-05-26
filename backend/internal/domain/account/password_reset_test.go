package account

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/account/notification"
	"jcourse/internal/domain/account/verification"
	domainemail "jcourse/internal/domain/email"
	"jcourse/internal/domain/task"
)

func TestPasswordResetService_SendResetCodeSuccess(t *testing.T) {
	username := mustUsernameFromEmail(t, "alice@example.edu")
	repo := identity.NewMockRepository(map[string]*identity.Account{
		username: {ID: 1, Username: username},
	})
	codes := verification.NewMockCodeRepository()
	enqueuer := &resetFakeEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	svc := NewPasswordResetService(repo, codes, hasher, testUsernameDeriver(), verification.Config{
		CodeInterval: time.Minute,
		CodeTTL:      10 * time.Minute,
	})

	if err := svc.SendResetCode(context.Background(), "alice@example.edu"); err != nil {
		t.Fatalf("SendResetCode: %v", err)
	}
	assertResetVerificationEmailTask(t, enqueuer.tasks[0], "alice@example.edu")
	saved := codes.Saved["alice@example.edu"]
	if len(saved.Code) != 6 {
		t.Fatalf("saved code = %q, want 6 digits", saved.Code)
	}
}

func TestPasswordResetService_SendResetCodeRejectsUnknownUser(t *testing.T) {
	repo := identity.NewMockRepository(nil)
	svc := NewPasswordResetService(repo, verification.NewMockCodeRepository(), nil, testUsernameDeriver(), verification.Config{})

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
	svc := NewPasswordResetService(repo, verification.NewMockCodeRepository(), nil, testUsernameDeriver(), verification.Config{
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
	svc := NewPasswordResetService(repo, codes, hasher, testUsernameDeriver(), verification.Config{})

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
	svc := NewPasswordResetService(repo, codes, credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1}), testUsernameDeriver(), verification.Config{})

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
	svc := NewPasswordResetService(repo, verification.NewMockCodeRepository(), nil, testUsernameDeriver(), verification.Config{})

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
	svc := NewPasswordResetService(identity.NewMockRepository(nil), codes, credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1}), testUsernameDeriver(), verification.Config{})

	err := svc.ResetPassword(context.Background(), "nobody@example.edu", "123456", "newpass")
	if !errors.Is(err, identity.ErrNotFound) {
		t.Fatalf("ResetPassword error = %v, want ErrUserNotFound", err)
	}
}

type resetFakeEnqueuer struct {
	tasks []task.Task
}

func (e *resetFakeEnqueuer) Enqueue(_ context.Context, taskItem task.Task, _ ...task.EnqueueOption) error {
	e.tasks = append(e.tasks, taskItem)
	return nil
}

func assertResetVerificationEmailTask(t *testing.T, taskItem task.Task, to string) {
	t.Helper()
	if taskItem == nil {
		t.Fatal("expected email task to be enqueued")
	}
	if got := taskItem.Type(); got != domainemail.TaskTypeSendEmail {
		t.Fatalf("task type = %q, want %q", got, domainemail.TaskTypeSendEmail)
	}
	var payload domainemail.SendEmailPayload
	if err := json.Unmarshal(taskItem.Payload(), &payload); err != nil {
		t.Fatalf("unmarshal email payload: %v", err)
	}
	if payload.EmailType != "verification_code" || payload.Email.To != to || payload.Email.Subject != notification.VerificationCodeEmailSubject || payload.Email.Body == "" {
		t.Fatalf("email payload = %+v", payload)
	}
}
