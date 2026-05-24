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
	sender := &fakeCodeSender{}
	svc := newAccountService(accountRepo, userRepo, newFakeCodeRepo(), sender)
	ctx := context.Background()

	if err := svc.SendRegisterCode(ctx, application.SendRegisterCodeCommand{Email: "alice@example.edu"}); err != nil {
		t.Fatalf("SendRegisterCode: %v", err)
	}
	registered, err := svc.Register(ctx, application.RegisterCommand{Email: "Alice@Example.EDU", Code: sender.code, Password: "secret"})
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
	codes.saved["alice@example.edu"] = account.VerificationCode{Email: "alice@example.edu", Code: "123456", ExpiresAt: time.Now().Add(time.Minute)}
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
	if accountRepo.touchCount != 1 {
		t.Fatalf("TouchLastSeen count = %d, want 1", accountRepo.touchCount)
	}
}

func TestAccountCommandService_LoginLockedAfterMaxAttempts(t *testing.T) {
	username := accountUsername(t, "alice@example.edu")
	accountRepo := newFakeAccountRepo(map[string]*account.Account{
		"alice@example.edu": {ID: 1, Username: username, PasswordHash: mustHash(t, "secret")},
	})
	userRepo := newFakeAuthUserRepo(map[int]*auth.User{1: {ID: 1, Role: auth.RoleUser}})
	attempts := &fakeLoginAttemptRepo{counts: map[string]int{"alice@example.edu": 5}}
	svc := application.NewAccountCommandService(
		accountRepo,
		auth.NewCurrentUserService(userRepo),
		newFakeCodeRepo(),
		newFakeCodeRepo(),
		&fakeCodeSender{},
		account.NewDjangoPBKDF2SHA256PasswordHasher(1),
		testUsernameDeriver(),
		application.AccountCommandConfig{EmailWhitelist: []string{"@example.edu"}, CodeInterval: time.Minute, CodeTTL: 10 * time.Minute},
		account.PasswordResetConfig{CodeInterval: time.Minute, CodeTTL: 10 * time.Minute},
		attempts,
		5,
		15*time.Minute,
	)

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
	sender := &fakeCodeSender{}
	svc := newAccountService(accountRepo, userRepo, newFakeCodeRepo(), sender)
	ctx := context.Background()

	if err := svc.SendResetCode(ctx, application.SendResetCodeCommand{Email: "alice@example.edu"}); err != nil {
		t.Fatalf("SendResetCode: %v", err)
	}
	if sender.email != "alice@example.edu" {
		t.Fatalf("sender email = %q, want alice@example.edu", sender.email)
	}

	if err := svc.ResetPassword(ctx, application.ResetPasswordCommand{Email: "alice@example.edu", Code: sender.code, NewPassword: "newpass"}); err != nil {
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
	attempts := &fakeLoginAttemptRepo{counts: map[string]int{}}
	svc := application.NewAccountCommandService(
		accountRepo,
		auth.NewCurrentUserService(userRepo),
		newFakeCodeRepo(),
		newFakeCodeRepo(),
		&fakeCodeSender{},
		account.NewDjangoPBKDF2SHA256PasswordHasher(1),
		testUsernameDeriver(),
		application.AccountCommandConfig{EmailWhitelist: []string{"@example.edu"}, CodeInterval: time.Minute, CodeTTL: 10 * time.Minute},
		account.PasswordResetConfig{CodeInterval: time.Minute, CodeTTL: 10 * time.Minute},
		attempts,
		3,
		15*time.Minute,
	)
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

func newAccountService(accountRepo *fakeAccountRepo, userRepo *fakeAuthUserRepo, codes *fakeCodeRepo, sender *fakeCodeSender) *application.AccountCommandService {
	accountRepo.userRepo = userRepo
	return application.NewAccountCommandService(
		accountRepo,
		auth.NewCurrentUserService(userRepo),
		codes,
		newFakeCodeRepo(),
		sender,
		account.NewDjangoPBKDF2SHA256PasswordHasher(1),
		testUsernameDeriver(),
		application.AccountCommandConfig{EmailWhitelist: []string{"@example.edu"}, CodeInterval: time.Minute, CodeTTL: 10 * time.Minute},
		account.PasswordResetConfig{CodeInterval: time.Minute, CodeTTL: 10 * time.Minute},
		&fakeLoginAttemptRepo{},
		5,
		15*time.Minute,
	)
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := account.NewDjangoPBKDF2SHA256PasswordHasher(1).Hash(password)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	return hash
}

type fakeAccountRepo struct {
	nextID             int
	accountsByID       map[int]*account.Account
	accountsByEmail    map[string]*account.Account
	accountsByUsername map[string]*account.Account
	userRepo           *fakeAuthUserRepo
	touchCount         int
}

func newFakeAccountRepo(accounts map[string]*account.Account) *fakeAccountRepo {
	if accounts == nil {
		accounts = map[string]*account.Account{}
	}
	repo := &fakeAccountRepo{nextID: 1, accountsByID: map[int]*account.Account{}, accountsByEmail: map[string]*account.Account{}, accountsByUsername: map[string]*account.Account{}}
	for email, acct := range accounts {
		copy := *acct
		if copy.Username == "" {
			copy.Username = accountUsernameFromEmail(email)
		}
		repo.accountsByEmail[email] = &copy
		repo.accountsByUsername[copy.Username] = &copy
		repo.accountsByID[copy.ID] = &copy
		if copy.ID >= repo.nextID {
			repo.nextID = copy.ID + 1
		}
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
	return account.NewBLAKE2bUsernameDeriver("SALT")
}

func (r *fakeAccountRepo) Create(_ context.Context, acct *account.Account) error {
	if _, ok := r.accountsByUsername[acct.Username]; ok {
		return errors.New("duplicate user")
	}
	copy := *acct
	copy.ID = r.nextID
	r.nextID++
	r.accountsByID[copy.ID] = &copy
	r.accountsByEmail[copy.Email] = &copy
	r.accountsByUsername[copy.Username] = &copy
	if r.userRepo != nil {
		r.userRepo.users[copy.ID] = &auth.User{ID: copy.ID, Role: auth.RoleUser}
	}
	acct.ID = copy.ID
	return nil
}

func (r *fakeAccountRepo) Update(_ context.Context, acct *account.Account) error {
	copy := *acct
	r.accountsByID[acct.ID] = &copy
	r.accountsByEmail[acct.Email] = &copy
	r.accountsByUsername[acct.Username] = &copy
	return nil
}

func (r *fakeAccountRepo) TouchLastSeen(_ context.Context, id int, at time.Time) error {
	r.touchCount++
	acct, ok := r.accountsByID[id]
	if !ok {
		return errors.New("account not found")
	}
	acct.LastSeenAt = at
	return nil
}

func (r *fakeAccountRepo) FindByID(_ context.Context, id int) (*account.Account, error) {
	acct, ok := r.accountsByID[id]
	if !ok {
		return nil, nil
	}
	copy := *acct
	return &copy, nil
}

func (r *fakeAccountRepo) FindByUsername(_ context.Context, username string) (*account.Account, error) {
	acct, ok := r.accountsByUsername[username]
	if !ok {
		return nil, nil
	}
	copy := *acct
	return &copy, nil
}

func (r *fakeAccountRepo) FindByEmail(_ context.Context, email string) (*account.Account, error) {
	acct, ok := r.accountsByEmail[email]
	if !ok {
		return nil, nil
	}
	copy := *acct
	return &copy, nil
}

type fakeAuthUserRepo struct {
	users map[int]*auth.User
}

func newFakeAuthUserRepo(users map[int]*auth.User) *fakeAuthUserRepo {
	if users == nil {
		users = map[int]*auth.User{}
	}
	return &fakeAuthUserRepo{users: users}
}

func (r *fakeAuthUserRepo) Update(_ context.Context, u *auth.User) error {
	copy := *u
	r.users[u.ID] = &copy
	return nil
}

func (r *fakeAuthUserRepo) FindByID(_ context.Context, id int) (*auth.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	copy := *u
	return &copy, nil
}

func (r *fakeAuthUserRepo) FindByRole(_ context.Context, role string) ([]auth.User, error) {
	users := make([]auth.User, 0)
	for _, u := range r.users {
		if u.Role == role {
			users = append(users, *u)
		}
	}
	return users, nil
}

type fakeCodeRepo struct {
	saved        map[string]account.VerificationCode
	cooldownTill map[string]time.Time
}

func newFakeCodeRepo() *fakeCodeRepo {
	return &fakeCodeRepo{saved: map[string]account.VerificationCode{}, cooldownTill: map[string]time.Time{}}
}

func (r *fakeCodeRepo) ReserveSend(_ context.Context, email string, interval time.Duration) (time.Duration, error) {
	now := time.Now()
	if till := r.cooldownTill[email]; till.After(now) {
		return time.Until(till), nil
	}
	r.cooldownTill[email] = now.Add(interval)
	return 0, nil
}

func (r *fakeCodeRepo) Save(_ context.Context, code account.VerificationCode, _ time.Duration) error {
	r.saved[code.Email] = code
	return nil
}

func (r *fakeCodeRepo) Get(_ context.Context, email string) (*account.VerificationCode, error) {
	code, ok := r.saved[email]
	if !ok {
		return nil, nil
	}
	return &code, nil
}

func (r *fakeCodeRepo) Delete(_ context.Context, email string) error {
	delete(r.saved, email)
	return nil
}

type fakeCodeSender struct {
	email string
	code  string
}

func (s *fakeCodeSender) SendVerificationCode(_ context.Context, email string, code string) error {
	s.email = email
	s.code = code
	return nil
}

type fakeEnqueuer struct{ enqueued bool }

func (f *fakeEnqueuer) Enqueue(context.Context, task.Task, ...task.EnqueueOption) error {
	f.enqueued = true
	return nil
}

type fakeLoginAttemptRepo struct {
	counts map[string]int
}

func (r *fakeLoginAttemptRepo) Increment(_ context.Context, email string) (int, error) {
	if r.counts == nil {
		r.counts = map[string]int{}
	}
	r.counts[email]++
	return r.counts[email], nil
}

func (r *fakeLoginAttemptRepo) Get(_ context.Context, email string) (int, error) {
	if r.counts == nil {
		r.counts = map[string]int{}
	}
	return r.counts[email], nil
}

func (r *fakeLoginAttemptRepo) Reset(_ context.Context, email string) error {
	if r.counts == nil {
		r.counts = map[string]int{}
	}
	r.counts[email] = 0
	return nil
}
