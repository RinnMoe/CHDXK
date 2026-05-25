package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeAdminUserRepo struct {
	users map[int]*User
}

func newFakeAdminUserRepo(users map[int]*User) *fakeAdminUserRepo {
	if users == nil {
		users = map[int]*User{}
	}
	return &fakeAdminUserRepo{users: users}
}

func (r *fakeAdminUserRepo) Update(ctx context.Context, u *User) error {
	copy := *u
	r.users[u.ID] = &copy
	return nil
}

func (r *fakeAdminUserRepo) FindByID(ctx context.Context, id int) (*User, error) {
	if u, ok := r.users[id]; ok {
		copy := *u
		return &copy, nil
	}
	return nil, nil
}

func (r *fakeAdminUserRepo) FindByRole(ctx context.Context, role string) ([]User, error) {
	var users []User
	for _, u := range r.users {
		if u.Role == role {
			users = append(users, *u)
		}
	}
	return users, nil
}

func TestAdminUserService_SuspendUserForDays(t *testing.T) {
	repo := newFakeAdminUserRepo(map[int]*User{2: {ID: 2, Role: RoleUser}})
	svc := NewAdminUserService(repo, AdminConfig{DefaultSuspendDays: 7})

	if err := svc.SuspendUserForDays(context.Background(), 1, 2, 0); err != nil {
		t.Fatalf("SuspendUserForDays: %v", err)
	}
	got := repo.users[2]
	if got.SuspendedAt == nil || got.SuspendTill == nil {
		t.Fatalf("user was not suspended: %+v", got)
	}
}

func TestAdminUserService_RejectsSelfOperation(t *testing.T) {
	svc := NewAdminUserService(newFakeAdminUserRepo(map[int]*User{1: {ID: 1, Role: RoleUser}}), AdminConfig{})

	err := svc.GrantAdmin(context.Background(), 1, 1)
	if !errors.Is(err, ErrCannotOperateSelf) {
		t.Fatalf("error = %v, want %v", err, ErrCannotOperateSelf)
	}
}

func TestAdminUserService_RejectsSuspendingAdmin(t *testing.T) {
	repo := newFakeAdminUserRepo(map[int]*User{2: {ID: 2, Role: RoleAdmin}})
	svc := NewAdminUserService(repo, AdminConfig{})

	err := svc.SuspendUserForDays(context.Background(), 1, 2, 1)
	if !errors.Is(err, ErrCannotSuspendAdmin) {
		t.Fatalf("error = %v, want %v", err, ErrCannotSuspendAdmin)
	}
}

func TestAdminUserService_GrantAndRevokeAdmin(t *testing.T) {
	repo := newFakeAdminUserRepo(map[int]*User{2: {ID: 2, Role: RoleUser}})
	svc := NewAdminUserService(repo, AdminConfig{})

	if err := svc.GrantAdmin(context.Background(), 1, 2); err != nil {
		t.Fatalf("GrantAdmin: %v", err)
	}
	if repo.users[2].Role != RoleAdmin {
		t.Fatalf("role after grant = %q, want %q", repo.users[2].Role, RoleAdmin)
	}
	if err := svc.RevokeAdmin(context.Background(), 1, 2); err != nil {
		t.Fatalf("RevokeAdmin: %v", err)
	}
	if repo.users[2].Role != RoleUser {
		t.Fatalf("role after revoke = %q, want %q", repo.users[2].Role, RoleUser)
	}
}
