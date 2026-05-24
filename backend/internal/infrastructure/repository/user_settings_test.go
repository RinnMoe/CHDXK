package repository_test

import (
	"context"
	"testing"

	"jcourse/internal/domain/setting"
	"jcourse/internal/infrastructure/repository"
)

func TestUserSettingsRepository_GetByUserID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserSettingsRepository(db)

	got, err := repo.GetByUserID(context.Background(), 999)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if got != nil {
		t.Fatalf("settings = %+v, want nil", got)
	}
}

func TestUserSettingsRepository_SaveAndGetByUserID(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserSettingsRepository(db)
	ctx := context.Background()
	user := seedUser(t, db)

	if err := repo.Save(ctx, &setting.UserSettings{UserID: user.ID, CurrentSemester: "2024-2025-2"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if got == nil || got.UserID != user.ID || got.CurrentSemester != "2024-2025-2" {
		t.Fatalf("settings = %+v", got)
	}

	if err := repo.Save(ctx, &setting.UserSettings{UserID: user.ID, CurrentSemester: "2025-2026-1"}); err != nil {
		t.Fatalf("Save update: %v", err)
	}
	got, err = repo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByUserID after update: %v", err)
	}
	if got.CurrentSemester != "2025-2026-1" {
		t.Fatalf("CurrentSemester = %q, want updated", got.CurrentSemester)
	}
}
