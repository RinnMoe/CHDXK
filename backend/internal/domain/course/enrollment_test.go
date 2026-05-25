package course

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeEnrollmentCourseRepo struct {
	courses map[int]*Course
	offered map[string]bool
}

func newFakeEnrollmentCourseRepo() *fakeEnrollmentCourseRepo {
	return &fakeEnrollmentCourseRepo{courses: map[int]*Course{}, offered: map[string]bool{}}
}

func (r *fakeEnrollmentCourseRepo) Get(ctx context.Context, courseID int) (*Course, error) {
	if c, ok := r.courses[courseID]; ok {
		copy := *c
		return &copy, nil
	}
	return nil, nil
}

func (r *fakeEnrollmentCourseRepo) OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error) {
	return r.offered[semester], nil
}

func (r *fakeEnrollmentCourseRepo) OfferedSemesterExists(ctx context.Context, semester string) (bool, error) {
	return r.offered[semester], nil
}

type fakeEnrollmentRepo struct {
	created *CourseEnrollment
}

func (r *fakeEnrollmentRepo) Create(ctx context.Context, enrollment *CourseEnrollment) error {
	copy := *enrollment
	r.created = &copy
	return nil
}

func (r *fakeEnrollmentRepo) SyncFromCoursePairs(ctx context.Context, userID int, semester string, pairs []CourseCodeTeacher) (int64, error) {
	return 0, nil
}

func (r *fakeEnrollmentRepo) Delete(ctx context.Context, enrollmentID, userID int) error {
	return nil
}

func TestEnrollmentService_CreateUsesLastSemester(t *testing.T) {
	courseRepo := newFakeEnrollmentCourseRepo()
	courseRepo.courses[1] = &Course{ID: 1, LastSemester: "2025-2026-1"}
	enrollmentRepo := &fakeEnrollmentRepo{}
	svc := NewEnrollmentService(courseRepo, enrollmentRepo)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	if err := svc.Create(context.Background(), 10, 1, " 2025-2026-1 ", now); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if enrollmentRepo.created == nil || enrollmentRepo.created.Semester != "2025-2026-1" || !enrollmentRepo.created.CreatedAt.Equal(now) {
		t.Fatalf("created enrollment = %+v", enrollmentRepo.created)
	}
}

func TestEnrollmentService_CreateUsesOfferedCourse(t *testing.T) {
	courseRepo := newFakeEnrollmentCourseRepo()
	courseRepo.courses[1] = &Course{ID: 1, LastSemester: "2025-2026-1"}
	courseRepo.offered["2024-2025-2"] = true
	enrollmentRepo := &fakeEnrollmentRepo{}
	svc := NewEnrollmentService(courseRepo, enrollmentRepo)

	if err := svc.Create(context.Background(), 10, 1, "2024-2025-2", time.Now()); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if enrollmentRepo.created == nil || enrollmentRepo.created.Semester != "2024-2025-2" {
		t.Fatalf("created enrollment = %+v", enrollmentRepo.created)
	}
}

func TestEnrollmentService_CreateRejectsMissingSemester(t *testing.T) {
	svc := NewEnrollmentService(newFakeEnrollmentCourseRepo(), &fakeEnrollmentRepo{})

	err := svc.Create(context.Background(), 10, 1, " ", time.Now())
	if !errors.Is(err, ErrSemesterRequired) {
		t.Fatalf("error = %v, want %v", err, ErrSemesterRequired)
	}
}

func TestEnrollmentService_CreateRejectsMissingOfferedCourse(t *testing.T) {
	courseRepo := newFakeEnrollmentCourseRepo()
	courseRepo.courses[1] = &Course{ID: 1, LastSemester: "2025-2026-1"}
	svc := NewEnrollmentService(courseRepo, &fakeEnrollmentRepo{})

	err := svc.Create(context.Background(), 10, 1, "2024-2025-2", time.Now())
	if !errors.Is(err, ErrOfferedCourseNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrOfferedCourseNotFound)
	}
}
