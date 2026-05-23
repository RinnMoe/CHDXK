package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/task"
)

func TestCurrentUserService_GetUserRejectsSuspendedUser(t *testing.T) {
	repo := newCurrentUserServiceFakeRepo()
	nowTime := time.Now()
	suspendedAt := nowTime.Add(-time.Hour)
	suspendTill := nowTime.Add(time.Hour)
	repo.users[1] = &User{ID: 1, SuspendedAt: &suspendedAt, SuspendTill: &suspendTill}

	_, err := NewCurrentUserService(repo).GetUser(context.Background(), 1)
	if !errors.Is(err, ErrUserSuspended) {
		t.Fatalf("GetUser error = %v, want ErrUserSuspended", err)
	}
}

func TestCurrentUserService_GetUserEnqueuesCleanupForExpiredSuspension(t *testing.T) {
	repo := newCurrentUserServiceFakeRepo()
	nowTime := time.Now()
	taskEnqueuer = &fakeEnqueuer{}
	oldEnqueuer := task.SetEnqueuerForTest(taskEnqueuer)
	t.Cleanup(func() { task.SetEnqueuer(oldEnqueuer) })
	suspendedAt := nowTime.Add(-2 * time.Hour)
	suspendTill := nowTime.Add(-time.Hour)
	repo.users[1] = &User{ID: 1, SuspendedAt: &suspendedAt, SuspendTill: &suspendTill}

	u, err := NewCurrentUserService(repo).GetUser(context.Background(), 1)
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

type currentUserServiceFakeRepo struct {
	users map[int]*User
}

func newCurrentUserServiceFakeRepo() *currentUserServiceFakeRepo {
	return &currentUserServiceFakeRepo{users: map[int]*User{}}
}

func (r *currentUserServiceFakeRepo) Update(_ context.Context, u *User) error {
	copy := *u
	r.users[u.ID] = &copy
	return nil
}

func (r *currentUserServiceFakeRepo) FindByID(_ context.Context, id int) (*User, error) {
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
