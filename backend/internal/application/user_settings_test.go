package application_test

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
)

type fakeUserSettingsRepo struct {
	settings map[int]*setting.UserSettings
}

type fakeSettingsCourseRepo struct {
	offeredSemesters map[string]bool
}

func newFakeUserSettingsRepo() *fakeUserSettingsRepo {
	return &fakeUserSettingsRepo{settings: make(map[int]*setting.UserSettings)}
}

func (r *fakeUserSettingsRepo) GetByUserID(ctx context.Context, userID int) (*setting.UserSettings, error) {
	if s, ok := r.settings[userID]; ok {
		copy := *s
		return &copy, nil
	}
	return nil, nil
}

func (r *fakeUserSettingsRepo) Save(ctx context.Context, settings *setting.UserSettings) error {
	copy := *settings
	r.settings[settings.UserID] = &copy
	return nil
}

func newFakeSettingsCourseRepo(semesters ...string) *fakeSettingsCourseRepo {
	repo := &fakeSettingsCourseRepo{offeredSemesters: make(map[string]bool)}
	for _, semester := range semesters {
		repo.offeredSemesters[semester] = true
	}
	return repo
}

func (r *fakeSettingsCourseRepo) Get(ctx context.Context, courseID int) (*course.Course, error) {
	return nil, nil
}

func (r *fakeSettingsCourseRepo) OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error) {
	return false, nil
}

func (r *fakeSettingsCourseRepo) OfferedSemesterExists(ctx context.Context, semester string) (bool, error) {
	return r.offeredSemesters[semester], nil
}

func newSettingsCourseQuery(semesters ...string) *fakeCourseQuery {
	items := make([]course.FilterItem, 0, len(semesters))
	for _, semester := range semesters {
		items = append(items, course.FilterItem{Name: semester, Count: 1})
	}
	q := newFakeCourseQuery()
	q.filters = &course.CourseFilters{Semesters: items}
	return q
}

func TestUserSettingsQueryService_Get_DefaultsToLatestSemester(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	courseQuery := newSettingsCourseQuery("2025-2026-1", "2024-2025-2")
	courseRepo := newFakeSettingsCourseRepo("2025-2026-1", "2024-2025-2")
	svc := application.NewUserSettingsQueryService(repo, courseQuery, courseRepo)

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
	repo.settings[1] = &setting.UserSettings{UserID: 1, CurrentSemester: "2024-2025-2"}
	courseQuery := newSettingsCourseQuery("2025-2026-1", "2024-2025-2")
	courseRepo := newFakeSettingsCourseRepo("2025-2026-1", "2024-2025-2")
	svc := application.NewUserSettingsQueryService(repo, courseQuery, courseRepo)

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
	repo.settings[1] = &setting.UserSettings{UserID: 1, CurrentSemester: "2020-2021-1"}
	courseQuery := newSettingsCourseQuery("2025-2026-1", "2024-2025-2")
	courseRepo := newFakeSettingsCourseRepo("2025-2026-1", "2024-2025-2")
	svc := application.NewUserSettingsQueryService(repo, courseQuery, courseRepo)

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
	courseRepo := newFakeSettingsCourseRepo("2025-2026-1", "2024-2025-2")
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
	courseRepo := newFakeSettingsCourseRepo("2025-2026-1")
	svc := application.NewUserSettingsCommandService(repo, courseRepo)

	_, err := svc.Update(context.Background(), 1, application.UpdateUserSettingsCommand{CurrentSemester: "2020-2021-1"})
	if !errors.Is(err, setting.ErrInvalidCurrentSemester) {
		t.Fatalf("error = %v, want %v", err, setting.ErrInvalidCurrentSemester)
	}
}

func TestUserSettingsCommandService_Update_AllowsExistingOfferedSemesterOutsideFilters(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	courseRepo := newFakeSettingsCourseRepo("2025-2026-1", "2024-2025-2")
	svc := application.NewUserSettingsCommandService(repo, courseRepo)

	dto, err := svc.Update(context.Background(), 1, application.UpdateUserSettingsCommand{CurrentSemester: "2024-2025-2"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if dto.CurrentSemester != "2024-2025-2" {
		t.Fatalf("CurrentSemester = %q, want offered semester", dto.CurrentSemester)
	}
}
