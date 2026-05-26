package course

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newFakeEnrollmentCourseRepo() *MockCourseRepository {
	repo := NewMockCourseRepository()
	repo.OnOfferedCourseExists = func(_ context.Context, _ int, semester string) (bool, error) {
		return repo.OfferedSemesters[semester], nil
	}
	return repo
}

func TestEnrollmentService_CreateUsesLastSemester(t *testing.T) {
	courseRepo := newFakeEnrollmentCourseRepo()
	courseRepo.Courses[1] = &CourseView{ID: 1, LastSemester: "2025-2026-1"}
	enrollmentRepo := &MockCourseEnrollmentRepository{}
	svc := NewEnrollmentService(courseRepo, enrollmentRepo)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	if err := svc.Create(context.Background(), 10, 1, " 2025-2026-1 ", now); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if enrollmentRepo.Created == nil || enrollmentRepo.Created.Semester != "2025-2026-1" || !enrollmentRepo.Created.CreatedAt.Equal(now) {
		t.Fatalf("created enrollment = %+v", enrollmentRepo.Created)
	}
}

func TestEnrollmentService_CreateUsesOfferedCourse(t *testing.T) {
	courseRepo := newFakeEnrollmentCourseRepo()
	courseRepo.Courses[1] = &CourseView{ID: 1, LastSemester: "2025-2026-1"}
	courseRepo.OfferedSemesters["2024-2025-2"] = true
	enrollmentRepo := &MockCourseEnrollmentRepository{}
	svc := NewEnrollmentService(courseRepo, enrollmentRepo)

	if err := svc.Create(context.Background(), 10, 1, "2024-2025-2", time.Now()); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if enrollmentRepo.Created == nil || enrollmentRepo.Created.Semester != "2024-2025-2" {
		t.Fatalf("created enrollment = %+v", enrollmentRepo.Created)
	}
}

func TestEnrollmentService_CreateRejectsMissingSemester(t *testing.T) {
	svc := NewEnrollmentService(newFakeEnrollmentCourseRepo(), &MockCourseEnrollmentRepository{})

	err := svc.Create(context.Background(), 10, 1, " ", time.Now())
	if !errors.Is(err, ErrSemesterRequired) {
		t.Fatalf("error = %v, want %v", err, ErrSemesterRequired)
	}
}

func TestEnrollmentService_CreateRejectsMissingOfferedCourse(t *testing.T) {
	courseRepo := newFakeEnrollmentCourseRepo()
	courseRepo.Courses[1] = &CourseView{ID: 1, LastSemester: "2025-2026-1"}
	svc := NewEnrollmentService(courseRepo, &MockCourseEnrollmentRepository{})

	err := svc.Create(context.Background(), 10, 1, "2024-2025-2", time.Now())
	if !errors.Is(err, ErrOfferedCourseNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrOfferedCourseNotFound)
	}
}
