package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/application"
	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/task"
)

func TestAccountCommandService_SendRegisterCode(t *testing.T) {
	accountRepo := newFakeAccountRepo(nil)
	userRepo := newFakeAuthUserRepo(nil)
	svc := newAccountService(accountRepo, userRepo, newFakeCodeRepo(), &fakeCodeSender{})

	err := svc.SendRegisterCode(context.Background(), application.SendRegisterCodeCommand{Email: "alice@example.edu"})
	if err != nil {
		t.Fatalf("SendRegisterCode: %v", err)
	}

	err = svc.SendRegisterCode(context.Background(), application.SendRegisterCodeCommand{Email: "alice@example.edu"})
	if !errors.Is(err, account.ErrVerificationTooSoon) {
		t.Fatalf("second SendRegisterCode error = %v, want ErrVerificationTooSoon", err)
	}
}

func TestAccountCommandService_SendRegisterCodeRejectsEmailOutsideWhitelist(t *testing.T) {
	svc := newAccountService(newFakeAccountRepo(nil), newFakeAuthUserRepo(nil), newFakeCodeRepo(), &fakeCodeSender{})

	err := svc.SendRegisterCode(context.Background(), application.SendRegisterCodeCommand{Email: "alice@example.com"})
	if !errors.Is(err, account.ErrEmailNotAllowed) {
		t.Fatalf("SendRegisterCode error = %v, want ErrEmailNotAllowed", err)
	}
}

func TestAccountCommandService_RegisterAndLogin(t *testing.T) {
	accountRepo := newFakeAccountRepo(nil)
	userRepo := newFakeAuthUserRepo(nil)
	codes := newFakeCodeRepo()
	sender := &fakeCodeSender{}
	svc := newAccountService(accountRepo, userRepo, codes, sender)
	ctx := context.Background()

	if err := svc.SendRegisterCode(ctx, application.SendRegisterCodeCommand{Email: "alice@example.edu"}); err != nil {
		t.Fatalf("SendRegisterCode: %v", err)
	}
	registered, err := svc.Register(ctx, application.RegisterCommand{Email: "Alice@Example.EDU", Code: codes.Saved["alice@example.edu"].Code, Password: "secret"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if registered.ID == 0 || registered.Email != "" || registered.Role != auth.RoleUser {
		t.Fatalf("registered user = %+v", registered)
	}
	if registered.Username != accountUsername(t, "alice@example.edu") {
		t.Fatalf("registered username = %q", registered.Username)
	}

	loggedIn, err := svc.Login(ctx, application.LoginCommand{Email: "alice@example.edu", Password: "secret"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if loggedIn.ID != registered.ID {
		t.Fatalf("Login ID = %d, want %d", loggedIn.ID, registered.ID)
	}
}

func TestAccountCommandService_RegisterRejectsWrongCode(t *testing.T) {
	codes := newFakeCodeRepo()
	codes.Saved["alice@example.edu"] = account.VerificationCode{Email: "alice@example.edu", Code: "123456", ExpiresAt: time.Now().Add(time.Minute)}
	svc := newAccountService(newFakeAccountRepo(nil), newFakeAuthUserRepo(nil), codes, &fakeCodeSender{})

	_, err := svc.Register(context.Background(), application.RegisterCommand{Email: "alice@example.edu", Code: "000000", Password: "secret"})
	if !errors.Is(err, account.ErrVerificationCodeInvalid) {
		t.Fatalf("Register error = %v, want ErrVerificationCodeInvalid", err)
	}
}

func TestAccountCommandService_LoginRejectsWrongPassword(t *testing.T) {
	username := accountUsername(t, "alice@example.edu")
	accountRepo := newFakeAccountRepo(map[string]*account.Account{
		"alice@example.edu": {ID: 1, Username: username, PasswordHash: mustHash(t, "secret")},
	})
	userRepo := newFakeAuthUserRepo(map[int]*auth.User{1: {ID: 1, Role: auth.RoleUser}})
	svc := newAccountService(accountRepo, userRepo, newFakeCodeRepo(), &fakeCodeSender{})

	_, err := svc.Login(context.Background(), application.LoginCommand{Email: "alice@example.edu", Password: "wrong"})
	if !errors.Is(err, account.ErrInvalidCredentials) {
		t.Fatalf("Login error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAccountCommandService_LoginRejectsSuspendedUser(t *testing.T) {
	now := time.Now()
	suspendedAt := now.Add(-time.Hour)
	suspendTill := now.Add(time.Hour)
	username := accountUsername(t, "alice@example.edu")
	accountRepo := newFakeAccountRepo(map[string]*account.Account{
		"alice@example.edu": {ID: 1, Username: username, PasswordHash: mustHash(t, "secret")},
	})
	userRepo := newFakeAuthUserRepo(map[int]*auth.User{1: {ID: 1, Role: auth.RoleUser, SuspendedAt: &suspendedAt, SuspendTill: &suspendTill}})
	svc := newAccountService(accountRepo, userRepo, newFakeCodeRepo(), &fakeCodeSender{})

	_, err := svc.Login(context.Background(), application.LoginCommand{Email: "alice@example.edu", Password: "secret"})
	if !errors.Is(err, auth.ErrUserSuspended) {
		t.Fatalf("Login error = %v, want ErrUserSuspended", err)
	}
}

func TestAccountCommandService_LoginAllowsExpiredSuspensionAndEnqueuesCleanup(t *testing.T) {
	now := time.Now()
	suspendedAt := now.Add(-2 * time.Hour)
	suspendTill := now.Add(-time.Hour)
	username := accountUsername(t, "alice@example.edu")
	accountRepo := newFakeAccountRepo(map[string]*account.Account{
		"alice@example.edu": {ID: 1, Username: username, PasswordHash: mustHash(t, "secret")},
	})
	userRepo := newFakeAuthUserRepo(map[int]*auth.User{1: {ID: 1, Role: auth.RoleUser, SuspendedAt: &suspendedAt, SuspendTill: &suspendTill}})
	enqueuer := &fakeEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newAccountService(accountRepo, userRepo, newFakeCodeRepo(), &fakeCodeSender{})

	loggedIn, err := svc.Login(context.Background(), application.LoginCommand{Email: "alice@example.edu", Password: "secret"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if loggedIn.ID != 1 {
		t.Fatalf("Login ID = %d, want 1", loggedIn.ID)
	}
	if !enqueuer.enqueued {
		t.Fatal("expected cleanup task to be enqueued")
	}
	if accountRepo.TouchCount != 1 {
		t.Fatalf("TouchLastSeen count = %d, want 1", accountRepo.TouchCount)
	}
}

func TestAccountCommandService_LoginLockedAfterMaxAttempts(t *testing.T) {
	username := accountUsername(t, "alice@example.edu")
	accountRepo := newFakeAccountRepo(map[string]*account.Account{
		"alice@example.edu": {ID: 1, Username: username, PasswordHash: mustHash(t, "secret")},
	})
	userRepo := newFakeAuthUserRepo(map[int]*auth.User{1: {ID: 1, Role: auth.RoleUser}})
	attempts := account.NewMockLoginAttemptRepository(map[string]int{"alice@example.edu": 5})
	svc := newAccountServiceWithAttempts(accountRepo, userRepo, newFakeCodeRepo(), newFakeCodeRepo(), &fakeCodeSender{}, 5, attempts)

	_, err := svc.Login(context.Background(), application.LoginCommand{Email: "alice@example.edu", Password: "secret"})
	if !errors.Is(err, account.ErrLoginLocked) {
		t.Fatalf("Login error = %v, want ErrLoginLocked", err)
	}
}

func TestAccountCommandService_SendResetCodeAndResetPassword(t *testing.T) {
	username := accountUsername(t, "alice@example.edu")
	accountRepo := newFakeAccountRepo(map[string]*account.Account{
		"alice@example.edu": {ID: 1, Username: username, PasswordHash: mustHash(t, "oldpass")},
	})
	userRepo := newFakeAuthUserRepo(map[int]*auth.User{1: {ID: 1, Role: auth.RoleUser}})
	codes := newFakeCodeRepo()
	sender := &fakeCodeSender{}
	svc := newAccountServiceWithAttempts(accountRepo, userRepo, newFakeCodeRepo(), codes, sender, 5, account.NewMockLoginAttemptRepository(nil))
	ctx := context.Background()

	if err := svc.SendResetCode(ctx, application.SendResetCodeCommand{Email: "alice@example.edu"}); err != nil {
		t.Fatalf("SendResetCode: %v", err)
	}
	if sender.email.To != "alice@example.edu" {
		t.Fatalf("sender email.To = %q, want alice@example.edu", sender.email.To)
	}

	if err := svc.ResetPassword(ctx, application.ResetPasswordCommand{Email: "alice@example.edu", Code: codes.Saved["alice@example.edu"].Code, NewPassword: "newpass"}); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	loggedIn, err := svc.Login(ctx, application.LoginCommand{Email: "alice@example.edu", Password: "newpass"})
	if err != nil {
		t.Fatalf("Login after reset: %v", err)
	}
	if loggedIn.ID != 1 {
		t.Fatalf("Login ID = %d, want 1", loggedIn.ID)
	}
}

func TestAccountCommandService_SendResetCodeRejectsUnknownEmail(t *testing.T) {
	svc := newAccountService(newFakeAccountRepo(nil), newFakeAuthUserRepo(nil), newFakeCodeRepo(), &fakeCodeSender{})

	err := svc.SendResetCode(context.Background(), application.SendResetCodeCommand{Email: "nobody@example.edu"})
	if !errors.Is(err, account.ErrUserNotFound) {
		t.Fatalf("SendResetCode error = %v, want ErrUserNotFound", err)
	}
}

func TestAccountCommandService_ResetPasswordRejectsWrongCode(t *testing.T) {
	username := accountUsername(t, "alice@example.edu")
	accountRepo := newFakeAccountRepo(map[string]*account.Account{
		"alice@example.edu": {ID: 1, Username: username, PasswordHash: mustHash(t, "oldpass")},
	})
	userRepo := newFakeAuthUserRepo(map[int]*auth.User{1: {ID: 1, Role: auth.RoleUser}})
	svc := newAccountService(accountRepo, userRepo, newFakeCodeRepo(), &fakeCodeSender{})

	err := svc.ResetPassword(context.Background(), application.ResetPasswordCommand{Email: "alice@example.edu", Code: "000000", NewPassword: "newpass"})
	if !errors.Is(err, account.ErrVerificationCodeInvalid) {
		t.Fatalf("ResetPassword error = %v, want ErrVerificationCodeInvalid", err)
	}
}

func TestAccountCommandService_LoginFailedIncrementsAndLocks(t *testing.T) {
	username := accountUsername(t, "alice@example.edu")
	accountRepo := newFakeAccountRepo(map[string]*account.Account{
		"alice@example.edu": {ID: 1, Username: username, PasswordHash: mustHash(t, "secret")},
	})
	userRepo := newFakeAuthUserRepo(map[int]*auth.User{1: {ID: 1, Role: auth.RoleUser}})
	attempts := account.NewMockLoginAttemptRepository(map[string]int{})
	svc := newAccountServiceWithAttempts(accountRepo, userRepo, newFakeCodeRepo(), newFakeCodeRepo(), &fakeCodeSender{}, 3, attempts)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		_, err := svc.Login(ctx, application.LoginCommand{Email: "alice@example.edu", Password: "wrong"})
		if !errors.Is(err, account.ErrInvalidCredentials) {
			t.Fatalf("attempt %d error = %v, want ErrInvalidCredentials", i, err)
		}
	}

	_, err := svc.Login(ctx, application.LoginCommand{Email: "alice@example.edu", Password: "secret"})
	if !errors.Is(err, account.ErrLoginLocked) {
		t.Fatalf("error after 3 failures = %v, want ErrLoginLocked", err)
	}
}

func newAccountService(accountRepo *account.MockAccountRepository, userRepo *auth.MockUserRepository, codes *account.MockVerificationCodeRepository, sender *fakeCodeSender) *application.AccountCommandService {
	return newAccountServiceWithAttempts(accountRepo, userRepo, codes, newFakeCodeRepo(), sender, 5, account.NewMockLoginAttemptRepository(nil))
}

func newAccountServiceWithAttempts(
	accountRepo *account.MockAccountRepository,
	userRepo *auth.MockUserRepository,
	registerCodes *account.MockVerificationCodeRepository,
	resetCodes *account.MockVerificationCodeRepository,
	sender *fakeCodeSender,
	maxLoginAttempts int,
	attempts account.LoginAttemptRepository,
) *application.AccountCommandService {
	accountRepo.AfterCreate = func(acct *account.Account) {
		userRepo.Users[acct.ID] = &auth.User{ID: acct.ID, Role: auth.RoleUser}
	}
	hasher := account.NewDjangoPBKDF2SHA256PasswordHasher(account.PasswordHashConfig{Iterations: 1})
	usernames := testUsernameDeriver()
	return application.NewAccountCommandService(
		account.NewRegistrationService(accountRepo, registerCodes, sender, hasher, usernames, testRegistrationConfig()),
		account.NewLoginService(accountRepo, hasher, attempts, usernames, testLoginConfig(maxLoginAttempts)),
		account.NewPasswordResetService(accountRepo, resetCodes, sender, hasher, usernames, testPasswordResetConfig()),
		auth.NewCurrentUserService(userRepo),
	)
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := account.NewDjangoPBKDF2SHA256PasswordHasher(account.PasswordHashConfig{Iterations: 1}).Hash(password)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	return hash
}

func testRegistrationConfig() account.RegistrationConfig {
	return account.RegistrationConfig{
		EmailWhitelist: []string{"@example.edu"},
		CodeInterval:   time.Minute,
		CodeTTL:        10 * time.Minute,
		CodeLength:     6,
	}
}

func testPasswordResetConfig() account.PasswordResetConfig {
	return account.PasswordResetConfig{
		CodeInterval: time.Minute,
		CodeTTL:      10 * time.Minute,
		CodeLength:   6,
	}
}

func testLoginConfig(maxLoginAttempts int) account.LoginConfig {
	return account.LoginConfig{
		MaxAttempts: maxLoginAttempts,
		Lockout:     15 * time.Minute,
	}
}

func newFakeAccountRepo(accounts map[string]*account.Account) *account.MockAccountRepository {
	if accounts == nil {
		accounts = map[string]*account.Account{}
	}
	repo := account.NewMockAccountRepository(nil)
	for email, acct := range accounts {
		copy := *acct
		if copy.Username == "" {
			copy.Username = accountUsernameFromEmail(email)
		}
		repo.PutAccount(email, &copy)
	}
	return repo
}

func accountUsername(t *testing.T, email string) string {
	t.Helper()
	username, err := testUsernameDeriver().UsernameFromEmail(email)
	if err != nil {
		t.Fatalf("UsernameFromEmail: %v", err)
	}
	return username
}

func accountUsernameFromEmail(email string) string {
	username, err := testUsernameDeriver().UsernameFromEmail(email)
	if err != nil {
		return email
	}
	return username
}

func testUsernameDeriver() account.UsernameDeriver {
	return account.NewBLAKE2bUsernameDeriver(account.UsernameDeriverConfig{Salt: "SALT"})
}

func newFakeAuthUserRepo(users map[int]*auth.User) *auth.MockUserRepository {
	return auth.NewMockUserRepository(users)
}

func newFakeCodeRepo() *account.MockVerificationCodeRepository {
	return account.NewMockVerificationCodeRepository()
}

type fakeCodeSender struct {
	email account.Email
}

func (s *fakeCodeSender) SendEmail(_ context.Context, email account.Email) error {
	s.email = email
	return nil
}

type fakeEnqueuer struct{ enqueued bool }

func (f *fakeEnqueuer) Enqueue(context.Context, task.Task, ...task.EnqueueOption) error {
	f.enqueued = true
	return nil
}
