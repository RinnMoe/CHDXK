package application

import (
	"context"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/review"
)

type CourseReviewFilter struct{}
type UserReviewFilter struct{}

type ReviewQueryService struct {
	repo review.ReviewQuery
}

func NewReviewQueryService(repo review.ReviewQuery) *ReviewQueryService {
	return &ReviewQueryService{repo: repo}
}

func (s *ReviewQueryService) GetReviewsByCourse(ctx context.Context, courseID int, filter CourseReviewFilter) ([]ReviewDTO, error) {
	reviewFilter := review.ReviewFilter{
		CourseID: courseID,
	}

	reviews, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}

	views := make([]ReviewDTO, len(reviews))
	for i, r := range reviews {
		views[i] = newReviewDTO(&r)
	}

	return views, nil
}

func (s *ReviewQueryService) GetReviewsByUser(ctx context.Context, userID int, filter UserReviewFilter) ([]ReviewDTO, error) {
	reviewFilter := review.ReviewFilter{
		UserID: userID,
	}

	reviews, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}

	views := make([]ReviewDTO, len(reviews))
	for i, r := range reviews {
		views[i] = newReviewDTO(&r)
	}

	return views, nil
}

func (s *ReviewQueryService) GetReview(ctx context.Context, u *auth.User, reviewID int) (*ReviewDTO, error) {
	reviews, err := s.repo.FindBy(ctx, review.ReviewFilter{ReviewID: reviewID})
	if err != nil {
		return nil, err
	}
	view := newReviewDTO(&reviews[0])
	return &view, nil
}
