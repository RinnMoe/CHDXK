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
	repo.Users[1] = &User{ID: 1, SuspendedAt: &suspendedAt, SuspendTill: &suspendTill}

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
	repo.Users[1] = &User{ID: 1, SuspendedAt: &suspendedAt, SuspendTill: &suspendTill}

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

func TestCurrentUserService_SuspendUserSuspendsRegularUser(t *testing.T) {
	repo := newCurrentUserServiceFakeRepo()
	repo.Users[1] = &User{ID: 1, Role: RoleUser}

	if err := NewCurrentUserService(repo).SuspendUser(context.Background(), 1, 2*time.Hour); err != nil {
		t.Fatalf("SuspendUser: %v", err)
	}
	got := repo.Users[1]
	if got.SuspendedAt == nil || got.SuspendTill == nil {
		t.Fatalf("user was not suspended: %+v", got)
	}
	if d := got.SuspendTill.Sub(*got.SuspendedAt); d < 2*time.Hour-time.Second || d > 2*time.Hour+time.Second {
		t.Fatalf("suspension duration = %s, want about 2h", d)
	}
}

func TestCurrentUserService_SuspendUserSkipsAdmin(t *testing.T) {
	repo := newCurrentUserServiceFakeRepo()
	repo.Users[1] = &User{ID: 1, Role: RoleAdmin}

	if err := NewCurrentUserService(repo).SuspendUser(context.Background(), 1, 2*time.Hour); err != nil {
		t.Fatalf("SuspendUser: %v", err)
	}
	got := repo.Users[1]
	if got.SuspendedAt != nil || got.SuspendTill != nil {
		t.Fatalf("admin was suspended: %+v", got)
	}
}

func TestUserRoleHelpers(t *testing.T) {
	system := &User{Role: RoleSystem}
	if !system.IsSystemAPIKey() {
		t.Fatal("expected system role to be system api key")
	}
	if system.IsAdmin() {
		t.Fatal("system api key should not be admin")
	}
}

func newCurrentUserServiceFakeRepo() *MockUserRepository {
	return NewMockUserRepository(nil)
}

type fakeEnqueuer struct{ enqueued bool }

func (f *fakeEnqueuer) Enqueue(context.Context, task.Task, ...task.EnqueueOption) error {
	f.enqueued = true
	return nil
}

var taskEnqueuer *fakeEnqueuer
