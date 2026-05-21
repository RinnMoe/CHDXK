package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/task"
)

func TestAuthCommandService_SendRegisterCode(t *testing.T) {
	repo := newFakeUserRepo()
	codes := newFakeCodeRepo()
	sender := &fakeCodeSender{}
	svc := newAuthService(repo, codes, sender)

	err := svc.SendRegisterCode(context.Background(), application.SendRegisterCodeCommand{Email: "alice@example.edu"})
	if err != nil {
		t.Fatalf("SendRegisterCode: %v", err)
	}
	if sender.email != "alice@example.edu" || len(sender.code) != 6 {
		t.Fatalf("sender got email=%q code=%q", sender.email, sender.code)
	}

	err = svc.SendRegisterCode(context.Background(), application.SendRegisterCodeCommand{Email: "alice@example.edu"})
	if !errors.Is(err, auth.ErrVerificationTooSoon) {
		t.Fatalf("second SendRegisterCode error = %v, want ErrVerificationTooSoon", err)
	}
}

func TestAuthCommandService_SendRegisterCodeRejectsEmailOutsideWhitelist(t *testing.T) {
	svc := newAuthService(newFakeUserRepo(), newFakeCodeRepo(), &fakeCodeSender{})

	err := svc.SendRegisterCode(context.Background(), application.SendRegisterCodeCommand{Email: "alice@example.com"})
	if !errors.Is(err, auth.ErrEmailNotAllowed) {
		t.Fatalf("SendRegisterCode error = %v, want ErrEmailNotAllowed", err)
	}
}

func TestAuthCommandService_RegisterAndLogin(t *testing.T) {
	repo := newFakeUserRepo()
	codes := newFakeCodeRepo()
	sender := &fakeCodeSender{}
	svc := newAuthService(repo, codes, sender)
	ctx := context.Background()

	if err := svc.SendRegisterCode(ctx, application.SendRegisterCodeCommand{Email: "alice@example.edu"}); err != nil {
		t.Fatalf("SendRegisterCode: %v", err)
	}
	registered, err := svc.Register(ctx, application.RegisterCommand{Email: "Alice@Example.EDU", Code: sender.code, Password: "secret"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if registered.ID == 0 || registered.Email != "alice@example.edu" || registered.Role != auth.RoleUser {
		t.Fatalf("registered user = %+v", registered)
	}

	loggedIn, err := svc.Login(ctx, application.LoginCommand{Email: "alice@example.edu", Password: "secret"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if loggedIn.ID != registered.ID {
		t.Fatalf("Login ID = %d, want %d", loggedIn.ID, registered.ID)
	}
}

func TestAuthCommandService_RegisterRejectsWrongCode(t *testing.T) {
	codes := newFakeCodeRepo()
	codes.saved["alice@example.edu"] = auth.VerificationCode{
		Email:     "alice@example.edu",
		Code:      "123456",
		ExpiresAt: time.Now().Add(time.Minute),
	}
	svc := newAuthService(newFakeUserRepo(), codes, &fakeCodeSender{})

	_, err := svc.Register(context.Background(), application.RegisterCommand{Email: "alice@example.edu", Code: "000000", Password: "secret"})
	if !errors.Is(err, auth.ErrVerificationCodeInvalid) {
		t.Fatalf("Register error = %v, want ErrVerificationCodeInvalid", err)
	}
}

func TestAuthCommandService_LoginRejectsWrongPassword(t *testing.T) {
	repo := newFakeUserRepo()
	hasher := auth.NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, err := hasher.Hash("secret")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	repo.usersByEmail["alice@example.edu"] = &auth.User{ID: 1, Username: "alice@example.edu", Email: "alice@example.edu", Password: password, Role: auth.RoleUser}
	svc := newAuthService(repo, newFakeCodeRepo(), &fakeCodeSender{})

	_, err = svc.Login(context.Background(), application.LoginCommand{Email: "alice@example.edu", Password: "wrong"})
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("Login error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthCommandService_LoginRejectsSuspendedUser(t *testing.T) {
	repo := newFakeUserRepo()
	hasher := auth.NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, err := hasher.Hash("secret")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	now := time.Now()
	suspendedAt := now.Add(-time.Hour)
	suspendTill := now.Add(time.Hour)
	repo.usersByEmail["alice@example.edu"] = &auth.User{
		ID: 1, Username: "alice@example.edu", Email: "alice@example.edu", Password: password, Role: auth.RoleUser,
		SuspendedAt: &suspendedAt, SuspendTill: &suspendTill,
	}
	svc := newAuthService(repo, newFakeCodeRepo(), &fakeCodeSender{})

	_, err = svc.Login(context.Background(), application.LoginCommand{Email: "alice@example.edu", Password: "secret"})
	if !errors.Is(err, auth.ErrUserSuspended) {
		t.Fatalf("Login error = %v, want ErrUserSuspended", err)
	}
}

func TestAuthCommandService_LoginAllowsExpiredSuspensionAndEnqueuesCleanup(t *testing.T) {
	repo := newFakeUserRepo()
	hasher := auth.NewDjangoPBKDF2SHA256PasswordHasher(1)
	password, err := hasher.Hash("secret")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	now := time.Now()
	suspendedAt := now.Add(-2 * time.Hour)
	suspendTill := now.Add(-time.Hour)
	repo.usersByEmail["alice@example.edu"] = &auth.User{
		ID: 1, Username: "alice@example.edu", Email: "alice@example.edu", Password: password, Role: auth.RoleUser,
		SuspendedAt: &suspendedAt, SuspendTill: &suspendTill,
	}
	repo.usersByID[1] = repo.usersByEmail["alice@example.edu"]
	enqueuer := &fakeEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(enqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	svc := newAuthService(repo, newFakeCodeRepo(), &fakeCodeSender{})

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
	if repo.updateCount != 0 {
		t.Fatalf("Update count = %d, want 0", repo.updateCount)
	}
	if repo.touchCount != 1 {
		t.Fatalf("TouchLastSeen count = %d, want 1", repo.touchCount)
	}
}

func newAuthService(repo *fakeUserRepo, codes *fakeCodeRepo, sender *fakeCodeSender) *application.AuthCommandService {
	return application.NewAuthCommandService(
		repo,
		codes,
		sender,
		auth.NewDjangoPBKDF2SHA256PasswordHasher(1),
		application.AuthCommandConfig{
			EmailWhitelist: []string{"@example.edu"},
			CodeInterval:   time.Minute,
			CodeTTL:        10 * time.Minute,
		},
	)
}

type fakeUserRepo struct {
	nextID       int
	usersByID    map[int]*auth.User
	usersByEmail map[string]*auth.User
	updateCount  int
	touchCount   int
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{nextID: 1, usersByID: map[int]*auth.User{}, usersByEmail: map[string]*auth.User{}}
}

func (r *fakeUserRepo) Create(_ context.Context, u *auth.User) error {
	if _, ok := r.usersByEmail[u.Email]; ok {
		return errors.New("duplicate user")
	}
	copy := *u
	copy.ID = r.nextID
	r.nextID++
	r.usersByID[copy.ID] = &copy
	r.usersByEmail[copy.Email] = &copy
	u.ID = copy.ID
	return nil
}

func (r *fakeUserRepo) Update(_ context.Context, u *auth.User) error {
	r.updateCount++
	copy := *u
	r.usersByID[u.ID] = &copy
	r.usersByEmail[u.Email] = &copy
	return nil
}

func (r *fakeUserRepo) TouchLastSeen(_ context.Context, userID int, at time.Time) error {
	r.touchCount++
	u, ok := r.usersByID[userID]
	if !ok {
		return errors.New("user not found")
	}
	u.LastSeenAt = at
	return nil
}

func (r *fakeUserRepo) FindByID(_ context.Context, id int) (*auth.User, error) {
	u, ok := r.usersByID[id]
	if !ok {
		return nil, nil
	}
	copy := *u
	return &copy, nil
}

func (r *fakeUserRepo) FindByUsername(_ context.Context, username string) (*auth.User, error) {
	return r.FindByEmail(context.Background(), username)
}

func (r *fakeUserRepo) FindByEmail(_ context.Context, email string) (*auth.User, error) {
	u, ok := r.usersByEmail[email]
	if !ok {
		return nil, nil
	}
	copy := *u
	return &copy, nil
}

type fakeCodeRepo struct {
	saved        map[string]auth.VerificationCode
	cooldownTill map[string]time.Time
}

func newFakeCodeRepo() *fakeCodeRepo {
	return &fakeCodeRepo{saved: map[string]auth.VerificationCode{}, cooldownTill: map[string]time.Time{}}
}

func (r *fakeCodeRepo) ReserveSend(_ context.Context, email string, interval time.Duration) (time.Duration, error) {
	now := time.Now()
	if till := r.cooldownTill[email]; till.After(now) {
		return time.Until(till), nil
	}
	r.cooldownTill[email] = now.Add(interval)
	return 0, nil
}

func (r *fakeCodeRepo) Save(_ context.Context, code auth.VerificationCode, _ time.Duration) error {
	r.saved[code.Email] = code
	return nil
}

func (r *fakeCodeRepo) Get(_ context.Context, email string) (*auth.VerificationCode, error) {
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
