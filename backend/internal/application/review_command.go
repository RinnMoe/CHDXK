package application

import (
	"context"
	"errors"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type ReviewCommandService struct {
	courseRepo course.CourseRepository
	reviewRepo review.ReviewRepository
	policies   []review.CreatePolicy
}

func NewReviewCommandService(
	courseRepo course.CourseRepository,
	reviewRepo review.ReviewRepository,
	policies []review.CreatePolicy) *ReviewCommandService {
	return &ReviewCommandService{
		courseRepo: courseRepo,
		reviewRepo: reviewRepo,
		policies:   policies,
	}
}

func (s *ReviewCommandService) CreateReview(ctx context.Context, u *auth.User, cmd *CreateReviewCmd) error {
	c, err := s.courseRepo.Get(ctx, cmd.CourseID)
	if err != nil {
		return err
	}

	exists, err := s.courseRepo.OfferedCourseExists(ctx, cmd.CourseID, cmd.Semester)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("offered course not found for the given semester")
	}

	r := review.Review{
		CourseID:  cmd.CourseID,
		Semester:  cmd.Semester,
		Rating:    cmd.Rating,
		Content:   cmd.Content,
		Score:     cmd.Score,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.Validate(); err != nil {
		return err
	}

	g := review.NewGuardian(u, &r)
	if !g.CanCreate(ctx) {
		return errors.New("user cannot create review")
	}

	for _, policy := range s.policies {
		if err := policy.CanCreate(ctx, u, c, &r); err != nil {
			return err
		}
	}

	err = s.reviewRepo.Create(ctx, &r)
	if err != nil {
		return err
	}

	return nil
}

func (s *ReviewCommandService) UpdateReview(ctx context.Context, u *auth.User, cmd *UpdateReviewCmd) error {
	r, err := s.reviewRepo.Get(ctx, cmd.ReviewID)
	if err != nil {
		return err
	}

	g := review.NewGuardian(u, r)
	if !g.CanUpdate(ctx) {
		return errors.New("user cannot update review")
	}

	rv := r.MakeRevision()

	r.Semester = cmd.Semester
	r.Rating = cmd.Rating
	r.Content = cmd.Content
	r.Score = cmd.Score
	r.UpdatedAt = time.Now()

	if err := r.Validate(); err != nil {
		return err
	}

	err = s.reviewRepo.Update(ctx, r, rv)
	if err != nil {
		return err
	}

	return nil
}

func (s *ReviewCommandService) DeleteReview(ctx context.Context, u *auth.User, reviewID int) error {
	r, err := s.reviewRepo.Get(ctx, reviewID)
	if err != nil {
		return err
	}

	g := review.NewGuardian(u, r)
	if !g.CanDelete(ctx) {
		return errors.New("user cannot delete review")
	}

	err = s.reviewRepo.Delete(ctx, r)
	if err != nil {
		return err
	}

	return nil
}
