package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/repository"
)

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

	u.Role = auth.RoleAdmin
	u.SuspendedAt = new(time.Now().Add(-time.Hour))
	u.SuspendTill = new(time.Now().Add(time.Hour))
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
