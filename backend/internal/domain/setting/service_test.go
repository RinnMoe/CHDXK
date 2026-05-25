package setting

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/domain/course"
)

type fakeUserSettingsRepo struct {
	settings map[int]*UserSettings
}

func newFakeUserSettingsRepo() *fakeUserSettingsRepo {
	return &fakeUserSettingsRepo{settings: map[int]*UserSettings{}}
}

func (r *fakeUserSettingsRepo) GetByUserID(ctx context.Context, userID int) (*UserSettings, error) {
	if s, ok := r.settings[userID]; ok {
		copy := *s
		return &copy, nil
	}
	return nil, nil
}

func (r *fakeUserSettingsRepo) Save(ctx context.Context, settings *UserSettings) error {
	copy := *settings
	r.settings[settings.UserID] = &copy
	return nil
}

type fakeSettingsCourseRepo struct {
	offered map[string]bool
}

func newFakeSettingsCourseRepo(semesters ...string) *fakeSettingsCourseRepo {
	repo := &fakeSettingsCourseRepo{offered: map[string]bool{}}
	for _, semester := range semesters {
		repo.offered[semester] = true
	}
	return repo
}

func (r *fakeSettingsCourseRepo) OfferedSemesterExists(ctx context.Context, semester string) (bool, error) {
	return r.offered[semester], nil
}

func (r *fakeSettingsCourseRepo) Get(ctx context.Context, courseID int) (*course.Course, error) {
	return nil, nil
}

func (r *fakeSettingsCourseRepo) OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error) {
	return false, nil
}

func TestUserSettingsService_GetCurrentSemesterUsesSavedValidSemester(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	repo.settings[1] = &UserSettings{UserID: 1, CurrentSemester: "2024-2025-2"}
	svc := NewUserSettingsService(repo, newFakeSettingsCourseRepo("2025-2026-1", "2024-2025-2"))

	got, err := svc.GetCurrentSemester(context.Background(), 1, []string{"2025-2026-1", "2024-2025-2"})
	if err != nil {
		t.Fatalf("GetCurrentSemester: %v", err)
	}
	if got != "2024-2025-2" {
		t.Fatalf("current semester = %q, want saved", got)
	}
}

func TestUserSettingsService_GetCurrentSemesterFallsBackWhenSavedInvalid(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	repo.settings[1] = &UserSettings{UserID: 1, CurrentSemester: "2020-2021-1"}
	svc := NewUserSettingsService(repo, newFakeSettingsCourseRepo("2025-2026-1"))

	got, err := svc.GetCurrentSemester(context.Background(), 1, []string{"2025-2026-1"})
	if err != nil {
		t.Fatalf("GetCurrentSemester: %v", err)
	}
	if got != "2025-2026-1" {
		t.Fatalf("current semester = %q, want fallback", got)
	}
}

func TestUserSettingsService_UpdateCurrentSemester(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	svc := NewUserSettingsService(repo, newFakeSettingsCourseRepo("2024-2025-2"))

	settings, err := svc.UpdateCurrentSemester(context.Background(), 1, " 2024-2025-2 ")
	if err != nil {
		t.Fatalf("UpdateCurrentSemester: %v", err)
	}
	if settings.CurrentSemester != "2024-2025-2" {
		t.Fatalf("current semester = %q, want trimmed", settings.CurrentSemester)
	}
}

func TestUserSettingsService_UpdateCurrentSemesterRejectsInvalid(t *testing.T) {
	svc := NewUserSettingsService(newFakeUserSettingsRepo(), newFakeSettingsCourseRepo("2024-2025-2"))

	_, err := svc.UpdateCurrentSemester(context.Background(), 1, "2020-2021-1")
	if !errors.Is(err, ErrInvalidCurrentSemester) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidCurrentSemester)
	}
}
