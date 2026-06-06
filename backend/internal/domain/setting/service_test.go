package setting

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/domain/course"
)

func newFakeUserSettingsRepo() *MockRepository {
	return NewMockRepository()
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

func (r *fakeSettingsCourseRepo) Get(ctx context.Context, courseID int) (*course.CourseView, error) {
	return nil, nil
}

func (r *fakeSettingsCourseRepo) FindBy(ctx context.Context, filter course.CourseFilter) ([]course.CourseView, int64, error) {
	return nil, 0, nil
}

func (r *fakeSettingsCourseRepo) GetDetail(ctx context.Context, courseID int) (*course.CourseDetailView, error) {
	return nil, nil
}

func (r *fakeSettingsCourseRepo) FindOfferedCourses(ctx context.Context, courseID int) ([]course.OfferedCourseView, error) {
	return nil, nil
}

func (r *fakeSettingsCourseRepo) GetFilters(ctx context.Context) (*course.CourseFilters, error) {
	return nil, nil
}

func (r *fakeSettingsCourseRepo) UpdateModeratorRemark(ctx context.Context, courseID int, moderatorRemark string) error {
	return nil
}

func (r *fakeSettingsCourseRepo) RefreshRatingScores(ctx context.Context, config course.RatingScoreConfig) error {
	return nil
}

func (r *fakeSettingsCourseRepo) OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error) {
	return false, nil
}

func TestUserSettingsService_GetCurrentSemesterUsesSavedValidSemester(t *testing.T) {
	repo := newFakeUserSettingsRepo()
	repo.Settings[1] = &UserSettings{UserID: 1, CurrentSemester: "2024-2025-2"}
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
	repo.Settings[1] = &UserSettings{UserID: 1, CurrentSemester: "2020-2021-1"}
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

func TestSystemSettingsService_SaveRejectsUnregisteredKey(t *testing.T) {
	svc := NewSystemSettingsService(NewMockSystemRepository())

	_, err := svc.Save(context.Background(), "unknown", "value")
	if !errors.Is(err, ErrInvalidSystemSettingKey) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidSystemSettingKey)
	}
}

func TestSystemSettingsService_SaveRegisteredKeyWithoutValidator(t *testing.T) {
	repo := NewMockSystemRepository()
	svc := NewSystemSettingsService(repo)

	got, err := svc.Save(context.Background(), SystemSettingKeyCurrentSemester, "2025-2026-2")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got.Key != SystemSettingKeyCurrentSemester || got.Value != "2025-2026-2" {
		t.Fatalf("setting = %+v", got)
	}
}

func TestSystemSettingsService_SaveCurrentSemesterValidatesOfferedSemester(t *testing.T) {
	repo := NewMockSystemRepository()
	validators := NewSystemSettingValueValidatorFactory(newFakeSettingsCourseRepo("2025-2026-1"))
	svc := NewSystemSettingsService(repo, validators)

	got, err := svc.Save(context.Background(), SystemSettingKeyCurrentSemester, " 2025-2026-1 ")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got.Value != " 2025-2026-1 " {
		t.Fatalf("value = %q, want original", got.Value)
	}

	_, err = svc.Save(context.Background(), SystemSettingKeyCurrentSemester, "2020-2021-1")
	if !errors.Is(err, ErrInvalidCurrentSemester) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidCurrentSemester)
	}
}
