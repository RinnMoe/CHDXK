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

type fakeCourseRepo struct {
	courses        map[int]*course.Course
	offeredCourses map[int]map[string]bool
}

func (r *fakeCourseRepo) Get(ctx context.Context, courseID int) (*course.Course, error) {
	c, ok := r.courses[courseID]
	if !ok {
		return nil, nil
	}
	copy := *c
	return &copy, nil
}

func (r *fakeCourseRepo) OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error) {
	return r.offeredCourses[courseID][semester], nil
}

type fakeReviewRepo struct {
	nextID  int
	reviews map[int]*review.Review
}

func newFakeReviewRepo() *fakeReviewRepo {
	return &fakeReviewRepo{nextID: 1, reviews: map[int]*review.Review{}}
}

func (r *fakeReviewRepo) Create(ctx context.Context, rv *review.Review) error {
	copy := *rv
	copy.ID = r.nextID
	r.nextID++
	r.reviews[copy.ID] = &copy
	rv.ID = copy.ID
	return nil
}

func (r *fakeReviewRepo) Update(ctx context.Context, rv *review.Review, revision review.Revision) error {
	copy := *rv
	r.reviews[copy.ID] = &copy
	return nil
}

func (r *fakeReviewRepo) UpdateModeratorRemark(ctx context.Context, reviewID int, moderatorRemark string) error {
	return nil
}

func (r *fakeReviewRepo) Delete(ctx context.Context, rv *review.Review) error {
	delete(r.reviews, rv.ID)
	return nil
}

func (r *fakeReviewRepo) Get(ctx context.Context, reviewID int) (*review.Review, error) {
	rv, ok := r.reviews[reviewID]
	if !ok {
		return nil, nil
	}
	copy := *rv
	return &copy, nil
}

func TestServiceCreateAllowsCourseLastSemester(t *testing.T) {
	courseRepo := &fakeCourseRepo{
		courses: map[int]*course.Course{
			1: &course.Course{ID: 1, LastSemester: "2025-2026-1"},
		},
		offeredCourses: map[int]map[string]bool{},
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
	courseRepo := &fakeCourseRepo{
		courses: map[int]*course.Course{
			1: &course.Course{ID: 1, LastSemester: "2025-2026-1"},
		},
		offeredCourses: map[int]map[string]bool{},
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
	courseRepo := &fakeCourseRepo{
		courses: map[int]*course.Course{
			1: &course.Course{ID: 1, LastSemester: "2025-2026-1"},
		},
		offeredCourses: map[int]map[string]bool{},
	}
	reviewRepo := newFakeReviewRepo()
	reviewRepo.reviews[1] = &review.Review{ID: 1, CourseID: 1, UserID: 10, Semester: "2024-2025-2", Rating: 4, Content: "old"}
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
