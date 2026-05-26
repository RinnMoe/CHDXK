package application_test

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
)

func newFakeUserSettingsRepo() *setting.MockRepository {
	return setting.NewMockRepository()
}

func newSettingsCourseRepo(semesters ...string) *course.MockCourseRepository {
	items := make([]course.FilterItem, 0, len(semesters))
	repo := course.NewMockCourseRepository()
	for _, semester := range semesters {
		items = append(items, course.FilterItem{Name: semester, Count: 1})
		repo.OfferedSemesters[semester] = true
	}
	repo.Filters = &course.CourseFilters{Semesters: items}
	return repo
}

func TestUserSettingsQueryService_Get_DefaultsToLatestSemester(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	courseRepo := newSettingsCourseRepo("2025-2026-1", "2024-2025-2")
	svc := application.NewUserSettingsQueryService(repo, courseRepo)

	dto, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if dto.CurrentSemester != "2025-2026-1" {
		t.Fatalf("CurrentSemester = %q, want latest", dto.CurrentSemester)
	}
}

func TestUserSettingsQueryService_Get_UsesSavedValidSemester(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	repo.Settings[1] = &setting.UserSettings{UserID: 1, CurrentSemester: "2024-2025-2"}
	courseRepo := newSettingsCourseRepo("2025-2026-1", "2024-2025-2")
	svc := application.NewUserSettingsQueryService(repo, courseRepo)

	dto, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if dto.CurrentSemester != "2024-2025-2" {
		t.Fatalf("CurrentSemester = %q, want saved", dto.CurrentSemester)
	}
}

func TestUserSettingsQueryService_Get_FallsBackWhenSavedSemesterExpired(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	repo.Settings[1] = &setting.UserSettings{UserID: 1, CurrentSemester: "2020-2021-1"}
	courseRepo := newSettingsCourseRepo("2025-2026-1", "2024-2025-2")
	svc := application.NewUserSettingsQueryService(repo, courseRepo)

	dto, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if dto.CurrentSemester != "2025-2026-1" {
		t.Fatalf("CurrentSemester = %q, want fallback", dto.CurrentSemester)
	}
}

func TestUserSettingsCommandService_Update(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	courseRepo := newSettingsCourseRepo("2025-2026-1", "2024-2025-2")
	svc := application.NewUserSettingsCommandService(repo, courseRepo)

	dto, err := svc.Update(context.Background(), 1, application.UpdateUserSettingsCommand{CurrentSemester: " 2024-2025-2 "})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if dto.CurrentSemester != "2024-2025-2" {
		t.Fatalf("CurrentSemester = %q, want trimmed saved value", dto.CurrentSemester)
	}
	got, _ := repo.GetByUserID(context.Background(), 1)
	if got == nil || got.CurrentSemester != "2024-2025-2" {
		t.Fatalf("saved settings = %+v", got)
	}
}

func TestUserSettingsCommandService_Update_InvalidSemester(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	courseRepo := newSettingsCourseRepo("2025-2026-1")
	svc := application.NewUserSettingsCommandService(repo, courseRepo)

	_, err := svc.Update(context.Background(), 1, application.UpdateUserSettingsCommand{CurrentSemester: "2020-2021-1"})
	if !errors.Is(err, setting.ErrInvalidCurrentSemester) {
		t.Fatalf("error = %v, want %v", err, setting.ErrInvalidCurrentSemester)
	}
}

func TestUserSettingsCommandService_Update_AllowsExistingOfferedSemesterOutsideFilters(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	courseRepo := newSettingsCourseRepo("2025-2026-1", "2024-2025-2")
	svc := application.NewUserSettingsCommandService(repo, courseRepo)

	dto, err := svc.Update(context.Background(), 1, application.UpdateUserSettingsCommand{CurrentSemester: "2024-2025-2"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if dto.CurrentSemester != "2024-2025-2" {
		t.Fatalf("CurrentSemester = %q, want offered semester", dto.CurrentSemester)
	}
}
