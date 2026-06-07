package setting

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/domain/course"
)

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
