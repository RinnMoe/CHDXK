package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/repository"
)

func TestAccountRepository_CreateAndFind(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewAccountRepository(db)
	ctx := context.Background()

	acct := &identity.Account{
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "secret",
		CreatedAt:    time.Now(),
		LastSeenAt:   time.Now(),
	}
	if err := repo.Create(ctx, acct); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if acct.ID == 0 {
		t.Fatal("expected non-zero ID after Create")
	}

	got, err := repo.FindByID(ctx, acct.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Username != acct.Username || got.Email != acct.Email || got.PasswordHash != acct.PasswordHash {
		t.Fatalf("account = %+v, want username/email/password_hash from %+v", got, acct)
	}

	userRepo := repository.NewUserRepository(db)
	u, err := userRepo.FindByID(ctx, acct.ID)
	if err != nil {
		t.Fatalf("FindByID user: %v", err)
	}
	if u.Role != auth.RoleUser {
		t.Fatalf("created user role = %q, want %q", u.Role, auth.RoleUser)
	}
}

func TestAccountRepository_FindByUsernameAndEmail(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewAccountRepository(db)
	ctx := context.Background()
	e := seedUser(t, db)

	byUsername, err := repo.FindByUsername(ctx, e.Username)
	if err != nil {
		t.Fatalf("FindByUsername: %v", err)
	}
	if byUsername.ID != e.ID || byUsername.Username != e.Username {
		t.Fatalf("by username = %+v, want id=%d username=%q", byUsername, e.ID, e.Username)
	}

	byEmail, err := repo.FindByEmail(ctx, e.Email.String)
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if byEmail.ID != e.ID || byEmail.Email != e.Email.String {
		t.Fatalf("by email = %+v, want id=%d email=%q", byEmail, e.ID, e.Email.String)
	}
}

func TestAccountRepository_CreateWithoutEmail(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewAccountRepository(db)
	ctx := context.Background()

	acct := &identity.Account{
		Username:     "hashedusername",
		PasswordHash: "secret",
		CreatedAt:    time.Now(),
		LastSeenAt:   time.Now(),
	}
	if err := repo.Create(ctx, acct); err != nil {
		t.Fatalf("Create: %v", err)
	}

	var raw repository.UserEntity
	if err := db.First(&raw, acct.ID).Error; err != nil {
		t.Fatalf("load raw user: %v", err)
	}
	if raw.Email.Valid {
		t.Fatalf("stored email = %q, want NULL", raw.Email.String)
	}
}

func TestAccountRepository_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewAccountRepository(db)
	ctx := context.Background()

	got, err := repo.FindByEmail(ctx, "nobody@example.com")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if got != nil {
		t.Fatalf("FindByEmail got %+v, want nil", got)
	}
}

func TestAccountRepository_UpdatePasswordHash(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewAccountRepository(db)
	ctx := context.Background()
	e := seedUser(t, db)

	acct := &identity.Account{ID: e.ID, Username: e.Username, Email: e.Email.String, PasswordHash: "new_password", LastSeenAt: e.LastSeenAt.Add(time.Hour)}
	if err := repo.Update(ctx, acct); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.FindByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.PasswordHash != "new_password" {
		t.Fatalf("password_hash = %q, want new_password", got.PasswordHash)
	}
}

func TestUserRepository_FindByIDAndUpdateAuthFields(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()
	e := seedUser(t, db)

	u, err := repo.FindByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if u.ID != e.ID || u.Role != e.Role {
		t.Fatalf("user = %+v, want id=%d role=%q", u, e.ID, e.Role)
	}

	suspendedAt := time.Now().Add(-time.Hour)
	suspendTill := time.Now().Add(time.Hour)
	u.Role = auth.RoleAdmin
	u.SuspendedAt = &suspendedAt
	u.SuspendTill = &suspendTill
	if err := repo.Update(ctx, u); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.FindByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("FindByID after update: %v", err)
	}
	if got.Role != auth.RoleAdmin || got.SuspendedAt == nil || got.SuspendTill == nil {
		t.Fatalf("updated user = %+v", got)
	}

	got.ClearSuspension()
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update cleared suspension: %v", err)
	}
	got, err = repo.FindByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("FindByID after clear: %v", err)
	}
	if got.SuspendedAt != nil || got.SuspendTill != nil {
		t.Fatalf("expected suspension cleared, got %+v", got)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	got, err := repo.FindByID(ctx, 999999)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got != nil {
		t.Fatalf("FindByID got %+v, want nil", got)
	}
}
