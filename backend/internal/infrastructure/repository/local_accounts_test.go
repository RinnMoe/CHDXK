package repository_test

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
	"jcourse/internal/infrastructure/repository"
	"testing"
)

func TestIndependentAccountsAndSharedBinding(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := repository.NewUserIdentityRepository(db, "https://hub.test")
	initial, err := repo.Initialize(ctx)
	if err != nil || initial == nil {
		t.Fatal("initialize", err)
	}
	root, err := repo.FindCredential(ctx, initial.Username)
	if err != nil || root.User.ID != 1 || !root.User.IsSuperAdmin() || !root.User.MustChangePassword || bcrypt.CompareHashAndPassword([]byte(root.PasswordHash), []byte(initial.Password)) != nil {
		t.Fatal("invalid owner credential", err)
	}
	if repeat, err := repo.Initialize(ctx); err != nil || repeat != nil {
		t.Fatal("startup reset credential", err)
	}
	if _, err := repo.SubjectByUserID(ctx, 1); !errors.Is(err, point.ErrNoPointAccount) {
		t.Fatal("unbound owner points", err)
	}
	identity := &auth.ResHubIdentity{Issuer: "https://hub.test", Subject: "user:8", Username: "资料站昵称", IdentityRole: "superadmin", PermissionMode: "member", HasPointsAccount: true}
	accounts, err := repo.ResHubAccounts(ctx, identity, true)
	if err != nil || len(accounts) != 1 || accounts[0].ID == 1 || accounts[0].Role != auth.RoleAdmin {
		t.Fatal("first SSO account", accounts, err)
	}
	if err := repo.BindAccount(ctx, 1, identity); err != nil {
		t.Fatal(err)
	}
	accounts, err = repo.ResHubAccounts(ctx, identity, true)
	if err != nil || len(accounts) != 2 || accounts[0].Role != auth.RoleSuperAdmin {
		t.Fatal("shared binding", accounts, err)
	}
	for _, u := range accounts {
		if subject, err := repo.SubjectByUserID(ctx, u.ID); err != nil || subject != "user:8" {
			t.Fatal("shared point subject", subject, err)
		}
	}
	local, err := repo.CreateLocal(ctx, "local-user", "普通用户", "hash", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.BindAccount(ctx, local.ID, identity); err != nil {
		t.Fatal(err)
	}
	if err := repo.UnbindAccount(ctx, local.ID); err != nil {
		t.Fatal(err)
	}
	unbound, _ := repo.FindByID(ctx, local.ID)
	if unbound.Subject != "" || unbound.AuthVersion != local.AuthVersion+2 {
		t.Fatal("unbind retained authority")
	}
	if err := repo.SetPassword(ctx, local.ID, "changed-hash", true); err != nil {
		t.Fatal(err)
	}
	changed, _ := repo.FindByID(ctx, local.ID)
	if changed.AuthVersion != unbound.AuthVersion+1 || !changed.MustChangePassword {
		t.Fatal("password reset session version")
	}
	if err := repo.DeleteAccount(ctx, local.ID); err != nil {
		t.Fatal(err)
	}
	if removed, err := repo.FindCredential(ctx, "local-user"); err != nil || removed != nil {
		t.Fatal("deleted account can login")
	}
}
