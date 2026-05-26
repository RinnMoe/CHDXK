package review

import (
	"context"
	"errors"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
)

var (
	ErrReviewNotFound       = errors.New("review not found")
	ErrCourseNotFound       = errors.New("course not found")
	ErrOfferedCourseMissing = errors.New("offered course not found for the given semester")
	ErrUserCannotCreate     = errors.New("user cannot create review")
	ErrUserCannotUpdate     = errors.New("user cannot update review")
	ErrUserCannotModerate   = errors.New("user cannot update moderator remark")
	ErrUserCannotDelete     = errors.New("user cannot delete review")
)

type CreateReview struct {
	CourseID int
	Semester string
	UserID   int
	Rating   int
	Content  string
	Score    string
	Now      time.Time
}

type UpdateReview struct {
	ReviewID int
	Semester string
	Rating   int
	Content  string
	Score    string
	Now      time.Time
}

type UpdateModeratorRemark struct {
	ReviewID        int
	ModeratorRemark string
}

type Service struct {
	courseRepo course.CourseRepository
	reviewRepo ReviewRepository
	policies   []CreatePolicy
}

func NewService(courseRepo course.CourseRepository, reviewRepo ReviewRepository, policies []CreatePolicy) *Service {
	return &Service{courseRepo: courseRepo, reviewRepo: reviewRepo, policies: policies}
}

func (s *Service) Create(ctx context.Context, u *auth.User, cmd CreateReview) error {
	c, err := s.courseRepo.Get(ctx, cmd.CourseID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCourseNotFound
	}

	exists, err := s.semesterExists(ctx, cmd.CourseID, cmd.Semester, c.LastSemester)
	if err != nil {
		return err
	}
	if !exists {
		return ErrOfferedCourseMissing
	}

	r := Review{
		CourseID:  cmd.CourseID,
		Semester:  cmd.Semester,
		UserID:    cmd.UserID,
		Rating:    cmd.Rating,
		Content:   cmd.Content,
		Score:     cmd.Score,
		CreatedAt: cmd.Now,
		UpdatedAt: cmd.Now,
	}
	if err := r.Validate(); err != nil {
		return err
	}

	g := NewGuardian(u, &r)
	if !g.CanCreate(ctx) {
		return ErrUserCannotCreate
	}

	for _, policy := range s.policies {
		if err := policy.CanCreate(ctx, u, c, &r); err != nil {
			return err
		}
	}
	return s.reviewRepo.Create(ctx, &r)
}

func (s *Service) Update(ctx context.Context, u *auth.User, cmd UpdateReview) error {
	r, err := s.reviewRepo.Get(ctx, cmd.ReviewID)
	if err != nil {
		return err
	}
	if r == nil {
		return ErrReviewNotFound
	}

	g := NewGuardian(u, r)
	if !g.CanUpdate(ctx) {
		return ErrUserCannotUpdate
	}

	c, err := s.courseRepo.Get(ctx, r.CourseID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCourseNotFound
	}

	exists, err := s.semesterExists(ctx, r.CourseID, cmd.Semester, c.LastSemester)
	if err != nil {
		return err
	}
	if !exists {
		return ErrOfferedCourseMissing
	}

	rv := r.MakeRevision()
	r.ApplyUpdate(Update{
		Semester: cmd.Semester,
		Rating:   cmd.Rating,
		Content:  cmd.Content,
		Score:    cmd.Score,
		Now:      cmd.Now,
	})
	if err := r.Validate(); err != nil {
		return err
	}
	return s.reviewRepo.Update(ctx, r, rv)
}

func (s *Service) semesterExists(ctx context.Context, courseID int, semester string, lastSemester string) (bool, error) {
	if semester != "" && semester == lastSemester {
		return true, nil
	}
	return s.courseRepo.OfferedCourseExists(ctx, courseID, semester)
}

func (s *Service) Delete(ctx context.Context, u *auth.User, reviewID int) error {
	r, err := s.reviewRepo.Get(ctx, reviewID)
	if err != nil {
		return err
	}
	if r == nil {
		return ErrReviewNotFound
	}
	g := NewGuardian(u, r)
	if !g.CanDelete(ctx) {
		return ErrUserCannotDelete
	}
	return s.reviewRepo.Delete(ctx, r)
}

func (s *Service) UpdateModeratorRemark(ctx context.Context, u *auth.User, cmd UpdateModeratorRemark) error {
	r, err := s.reviewRepo.Get(ctx, cmd.ReviewID)
	if err != nil {
		return err
	}
	if r == nil {
		return ErrReviewNotFound
	}

	g := NewGuardian(u, r)
	if !g.CanUpdateModeratorRemark(ctx) {
		return ErrUserCannotModerate
	}
	r.ModeratorRemark = cmd.ModeratorRemark
	return s.reviewRepo.UpdateModeratorRemark(ctx, r.ID, r.ModeratorRemark)
}

type CreatePolicy interface {
	CanCreate(ctx context.Context, u *auth.User, c *course.CourseView, r *Review) error
}
