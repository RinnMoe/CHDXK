package application

import (
	"context"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/review"
)

type ReviewDTO struct {
	ID        int        `json:"id"`
	Course    *CourseDTO `json:"course,omitempty"`
	CourseID  int        `json:"course_id"`
	Grade     string     `json:"grade"`
	Rating    int        `json:"rating"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func newReviewDTO(r *review.ReviewForQuery) ReviewDTO {
	dto := ReviewDTO{
		ID:        r.ID,
		Course:    &CourseDTO{ID: r.CourseID},
		CourseID:  r.CourseID,
		Grade:     r.Grade,
		Rating:    r.Rating,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if r.Course != nil {
		dto.Course = new(newCourseDTO(r.Course))
	}
	return dto
}

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

	reviewDTOs := make([]ReviewDTO, len(reviews))
	for i, r := range reviews {
		reviewDTOs[i] = newReviewDTO(&r)
	}

	return reviewDTOs, nil
}

func (s *ReviewQueryService) GetReviewsByUser(ctx context.Context, userID int, filter UserReviewFilter) ([]ReviewDTO, error) {
	reviewFilter := review.ReviewFilter{
		UserID: userID,
	}

	reviews, err := s.repo.FindBy(ctx, reviewFilter)
	if err != nil {
		return nil, err
	}

	reviewDTOs := make([]ReviewDTO, len(reviews))
	for i, r := range reviews {
		reviewDTOs[i] = newReviewDTO(&r)
	}

	return reviewDTOs, nil
}

func (s *ReviewQueryService) GetReview(ctx context.Context, u *auth.User, reviewID int) (*ReviewDTO, error) {
	reviews, err := s.repo.FindBy(ctx, review.ReviewFilter{ReviewID: reviewID})
	if err != nil {
		return nil, err
	}
	dto := newReviewDTO(&reviews[0])
	return &dto, nil
}
