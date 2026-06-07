package auth

import (
	"context"
	"errors"
	"testing"
)

func TestAdminUserService_SuspendUserForDays(t *testing.T) {
	repo := NewMockUserRepository(map[int]*User{2: {ID: 2, Role: RoleUser}})
	svc := NewAdminUserService(repo)

	if err := svc.SuspendUserForDays(context.Background(), 1, 2, 7); err != nil {
		t.Fatalf("SuspendUserForDays: %v", err)
	}
	got := repo.Users[2]
	if got.SuspendedAt == nil || got.SuspendTill == nil {
		t.Fatalf("user was not suspended: %+v", got)
	}
}

func TestAdminUserService_RejectsSelfOperation(t *testing.T) {
	svc := NewAdminUserService(NewMockUserRepository(map[int]*User{1: {ID: 1, Role: RoleUser}}))

	err := svc.GrantAdmin(context.Background(), 1, 1)
	if !errors.Is(err, ErrCannotOperateSelf) {
		t.Fatalf("error = %v, want %v", err, ErrCannotOperateSelf)
	}
}

func TestAdminUserService_RejectsSuspendingAdmin(t *testing.T) {
	repo := NewMockUserRepository(map[int]*User{2: {ID: 2, Role: RoleAdmin}})
	svc := NewAdminUserService(repo)

	err := svc.SuspendUserForDays(context.Background(), 1, 2, 1)
	if !errors.Is(err, ErrCannotSuspendAdmin) {
		t.Fatalf("error = %v, want %v", err, ErrCannotSuspendAdmin)
	}
}

func TestAdminUserService_GrantAndRevokeAdmin(t *testing.T) {
	repo := NewMockUserRepository(map[int]*User{2: {ID: 2, Role: RoleUser}})
	svc := NewAdminUserService(repo)

	if err := svc.GrantAdmin(context.Background(), 1, 2); err != nil {
		t.Fatalf("GrantAdmin: %v", err)
	}
	if repo.Users[2].Role != RoleAdmin {
		t.Fatalf("role after grant = %q, want %q", repo.Users[2].Role, RoleAdmin)
	}
	if err := svc.RevokeAdmin(context.Background(), 1, 2); err != nil {
		t.Fatalf("RevokeAdmin: %v", err)
	}
	if repo.Users[2].Role != RoleUser {
		t.Fatalf("role after revoke = %q, want %q", repo.Users[2].Role, RoleUser)
	}
}

func TestAdminUserService_RejectsModifyingSuperAdmin(t *testing.T) {
	repo := NewMockUserRepository(map[int]*User{2: {ID: 2, Role: RoleSuperAdmin}})
	svc := NewAdminUserService(repo)

	err := svc.RevokeAdmin(context.Background(), 1, 2)
	if !errors.Is(err, ErrCannotModifySuperAdmin) {
		t.Fatalf("error = %v, want %v", err, ErrCannotModifySuperAdmin)
	}
}
