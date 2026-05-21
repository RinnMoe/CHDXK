package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/task"
)

func TestAuthService_GetUserRejectsSuspendedUser(t *testing.T) {
	repo := newAuthServiceFakeRepo()
	nowTime := time.Now()
	suspendedAt := nowTime.Add(-time.Hour)
	suspendTill := nowTime.Add(time.Hour)
	repo.users[1] = &User{ID: 1, Email: "alice@example.edu", SuspendedAt: &suspendedAt, SuspendTill: &suspendTill}

	_, err := NewAuthService(repo).GetUser(context.Background(), 1)
	if !errors.Is(err, ErrUserSuspended) {
		t.Fatalf("GetUser error = %v, want ErrUserSuspended", err)
	}
}

func TestAuthService_GetUserEnqueuesCleanupForExpiredSuspension(t *testing.T) {
	repo := newAuthServiceFakeRepo()
	nowTime := time.Now()
	taskEnqueuer = &fakeEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(taskEnqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	suspendedAt := nowTime.Add(-2 * time.Hour)
	suspendTill := nowTime.Add(-time.Hour)
	repo.users[1] = &User{ID: 1, Email: "alice@example.edu", SuspendedAt: &suspendedAt, SuspendTill: &suspendTill}

	u, err := NewAuthService(repo).GetUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.SuspendedAt == nil || u.SuspendTill == nil {
		t.Fatalf("expected expired suspension markers to remain until async cleanup, got %+v", u)
	}
	if !taskEnqueuer.enqueued {
		t.Fatal("expected clear suspension task to be enqueued")
	}
}

type authServiceFakeRepo struct {
	users map[int]*User
}

func newAuthServiceFakeRepo() *authServiceFakeRepo {
	return &authServiceFakeRepo{users: map[int]*User{}}
}

func (r *authServiceFakeRepo) Create(_ context.Context, u *User) error { r.users[u.ID] = u; return nil }

func (r *authServiceFakeRepo) Update(_ context.Context, u *User) error {
	copy := *u
	r.users[u.ID] = &copy
	return nil
}

func (r *authServiceFakeRepo) TouchLastSeen(_ context.Context, userID int, at time.Time) error {
	u, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	u.LastSeenAt = at
	return nil
}

func (r *authServiceFakeRepo) FindByID(_ context.Context, id int) (*User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	copy := *u
	return &copy, nil
}

type fakeEnqueuer struct{ enqueued bool }

func (f *fakeEnqueuer) Enqueue(context.Context, task.Task, ...task.EnqueueOption) error {
	f.enqueued = true
	return nil
}

var taskEnqueuer *fakeEnqueuer

func (r *authServiceFakeRepo) FindByUsername(_ context.Context, username string) (*User, error) {
	for _, u := range r.users {
		if u.Username == username {
			copy := *u
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *authServiceFakeRepo) FindByEmail(_ context.Context, email string) (*User, error) {
	for _, u := range r.users {
		if u.Email == email {
			copy := *u
			return &copy, nil
		}
	}
	return nil, nil
}
