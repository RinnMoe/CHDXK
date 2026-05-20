package repository_test

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/repository"
)

func TestUserRepository_Create(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	u := &auth.User{
		Username:   "alice",
		Email:      "alice@example.com",
		Role:       auth.RoleUser,
		Password:   "secret",
		CreatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	}

	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("expected non-zero ID after Create")
	}

	got, err := repo.FindByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.Username != u.Username {
		t.Errorf("Username: got %q, want %q", got.Username, u.Username)
	}
	if got.Email != u.Email {
		t.Errorf("Email: got %q, want %q", got.Email, u.Email)
	}
	if got.Role != u.Role {
		t.Errorf("Role: got %q, want %q", got.Role, u.Role)
	}
}

func TestUserRepository_FindByUsername(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	e := seedUser(t, db)

	got, err := repo.FindByUsername(ctx, e.Username)
	if err != nil {
		t.Fatalf("FindByUsername: %v", err)
	}
	if got.ID != e.ID {
		t.Errorf("ID: got %d, want %d", got.ID, e.ID)
	}
	if got.Username != e.Username {
		t.Errorf("Username: got %q, want %q", got.Username, e.Username)
	}
}

func TestUserRepository_FindByUsername_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.FindByUsername(ctx, "nobody")
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	e := seedUser(t, db)

	got, err := repo.FindByEmail(ctx, e.Email)
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if got.Email != e.Email {
		t.Errorf("Email: got %q, want %q", got.Email, e.Email)
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	e := seedUser(t, db)

	u := &auth.User{
		ID:         e.ID,
		Username:   e.Username,
		Email:      e.Email,
		Role:       auth.RoleAdmin,
		Password:   e.Password,
		CreatedAt:  e.CreatedAt,
		LastSeenAt: time.Now(),
	}

	if err := repo.Update(ctx, u); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.FindByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("FindByID after update: %v", err)
	}
	if got.Role != auth.RoleAdmin {
		t.Errorf("Role after update: got %q, want %q", got.Role, auth.RoleAdmin)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 999999)
	if err == nil {
		t.Fatal("expected error for missing ID, got nil")
	}
	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
}
