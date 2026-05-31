//go:build test

package application

import (
	"context"
	"encoding/json"
	"testing"

	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/task"
)

type captureEnqueuer struct {
	taskType string
	payload  []byte
}

func (e *captureEnqueuer) Enqueue(_ context.Context, t task.Task, _ ...task.EnqueueOption) error {
	e.taskType = t.Type()
	e.payload = t.Payload()
	return nil
}

func TestAdminUserCommandServiceResetPassword(t *testing.T) {
	ctx := context.Background()
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	accountRepo := identity.NewMockRepository(map[string]*identity.Account{})
	accountRepo.PutAccount("", &identity.Account{
		ID:           2,
		Username:     "alice",
		Email:        "alice@sjtu.edu.cn",
		PasswordHash: "old_hash",
	})
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{
		2: {ID: 2, Role: auth.RoleUser},
	})
	enqueuer := &captureEnqueuer{}
	prev := task.SetEnqueuerForTest(enqueuer)
	defer task.SetEnqueuerForTest(prev)

	svc := NewAdminUserCommandService(userRepo, accountRepo, hasher, AdminUserCommandConfig{})
	actor := &auth.User{ID: 1, Role: auth.RoleSuperAdmin}
	if err := svc.ResetPassword(ctx, actor, 2, "newpass"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	acct, err := accountRepo.FindByID(ctx, 2)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if acct.PasswordHash == "old_hash" || !hasher.Verify("newpass", acct.PasswordHash) {
		t.Fatalf("password hash was not updated: %q", acct.PasswordHash)
	}
	if enqueuer.taskType != audit.TaskTypeRecordLog {
		t.Fatalf("task type = %q, want %q", enqueuer.taskType, audit.TaskTypeRecordLog)
	}
	var payload audit.RecordLogPayload
	if err := json.Unmarshal(enqueuer.payload, &payload); err != nil {
		t.Fatalf("unmarshal audit payload: %v", err)
	}
	if payload.ActorUserID != actor.ID || payload.Action != audit.ActionUserPasswordReset || payload.TargetType != audit.TargetTypeUser || payload.TargetID != "2" {
		t.Fatalf("audit payload = %+v", payload)
	}
	if payload.Details["email"] != "alice@sjtu.edu.cn" || payload.Details["had_password_before"] != true {
		t.Fatalf("audit details = %+v", payload.Details)
	}
}

func TestAdminUserCommandServiceResetPasswordRejectsEmptyPassword(t *testing.T) {
	hasher := credential.NewDjangoPBKDF2SHA256PasswordHasher(credential.PasswordHashConfig{Iterations: 1})
	accountRepo := identity.NewMockRepository(map[string]*identity.Account{
		"alice@sjtu.edu.cn": {ID: 2, Username: "alice", Email: "alice@sjtu.edu.cn", PasswordHash: "old_hash"},
	})
	userRepo := auth.NewMockUserRepository(map[int]*auth.User{
		2: {ID: 2, Role: auth.RoleUser},
	})
	svc := NewAdminUserCommandService(userRepo, accountRepo, hasher, AdminUserCommandConfig{})

	err := svc.ResetPassword(context.Background(), &auth.User{ID: 1, Role: auth.RoleSuperAdmin}, 2, "  ")
	if err != credential.ErrPasswordRequired {
		t.Fatalf("error = %v, want %v", err, credential.ErrPasswordRequired)
	}
}
