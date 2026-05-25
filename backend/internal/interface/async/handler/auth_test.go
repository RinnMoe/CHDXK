package handler

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"

	"jcourse/internal/domain/auth"
)

func TestSuspendUserHandlerSuspendsUser(t *testing.T) {
	repo := auth.NewMockUserRepository(map[int]*auth.User{1: {ID: 1, Role: auth.RoleUser}})
	handler := NewSuspendUserHandler(auth.NewCurrentUserService(repo))
	task := auth.NewSuspendUserTask(1, 2*time.Hour)

	if err := handler.ProcessTask(context.Background(), asynq.NewTask(task.Type(), task.Payload())); err != nil {
		t.Fatalf("ProcessTask: %v", err)
	}
	got := repo.Users[1]
	if got.SuspendedAt == nil || got.SuspendTill == nil {
		t.Fatalf("user was not suspended: %+v", got)
	}
}

func TestSuspendUserHandlerSkipsAdmin(t *testing.T) {
	repo := auth.NewMockUserRepository(map[int]*auth.User{1: {ID: 1, Role: auth.RoleAdmin}})
	handler := NewSuspendUserHandler(auth.NewCurrentUserService(repo))
	task := auth.NewSuspendUserTask(1, 2*time.Hour)

	if err := handler.ProcessTask(context.Background(), asynq.NewTask(task.Type(), task.Payload())); err != nil {
		t.Fatalf("ProcessTask: %v", err)
	}
	got := repo.Users[1]
	if got.SuspendedAt != nil || got.SuspendTill != nil {
		t.Fatalf("admin was suspended: %+v", got)
	}
}
