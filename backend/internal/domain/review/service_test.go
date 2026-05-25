package review_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

func newFakeReviewRepo() *review.MockReviewRepository {
	return review.NewMockReviewRepository()
}

func TestServiceCreateAllowsCourseLastSemester(t *testing.T) {
	courseRepo := &review.MockCourseRepository{
		Courses: map[int]*course.Course{
			1: &course.Course{ID: 1, LastSemester: "2025-2026-1"},
		},
		OfferedCourses: map[int]map[string]bool{},
	}
	reviewRepo := newFakeReviewRepo()
	svc := review.NewService(courseRepo, reviewRepo, nil)

	err := svc.Create(context.Background(), &auth.User{ID: 10}, review.CreateReview{
		CourseID: 1,
		Semester: "2025-2026-1",
		UserID:   10,
		Rating:   5,
		Content:  "good course",
		Now:      time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestServiceCreateRejectsMissingSemester(t *testing.T) {
	courseRepo := &review.MockCourseRepository{
		Courses: map[int]*course.Course{
			1: &course.Course{ID: 1, LastSemester: "2025-2026-1"},
		},
		OfferedCourses: map[int]map[string]bool{},
	}
	svc := review.NewService(courseRepo, newFakeReviewRepo(), nil)

	err := svc.Create(context.Background(), &auth.User{ID: 10}, review.CreateReview{
		CourseID: 1,
		Semester: "2024-2025-2",
		UserID:   10,
		Rating:   5,
		Content:  "good course",
		Now:      time.Now(),
	})
	if !errors.Is(err, review.ErrOfferedCourseMissing) {
		t.Fatalf("Create error = %v, want %v", err, review.ErrOfferedCourseMissing)
	}
}

func TestServiceUpdateAllowsCourseLastSemester(t *testing.T) {
	courseRepo := &review.MockCourseRepository{
		Courses: map[int]*course.Course{
			1: &course.Course{ID: 1, LastSemester: "2025-2026-1"},
		},
		OfferedCourses: map[int]map[string]bool{},
	}
	reviewRepo := newFakeReviewRepo()
	reviewRepo.Reviews[1] = &review.Review{ID: 1, CourseID: 1, UserID: 10, Semester: "2024-2025-2", Rating: 4, Content: "old"}
	svc := review.NewService(courseRepo, reviewRepo, nil)

	err := svc.Update(context.Background(), &auth.User{ID: 10}, review.UpdateReview{
		ReviewID: 1,
		Semester: "2025-2026-1",
		Rating:   5,
		Content:  "updated",
		Now:      time.Now(),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}
