package application

import (
	"time"

	"jcourse/internal/domain/review"
)

type ReviewDTO struct {
	ID        int             `json:"id"`
	Course    *CourseListItem `json:"course,omitempty"`
	CourseID  int             `json:"course_id"`
	Grade     string          `json:"grade"`
	Rating    int             `json:"rating"`
	Content   string          `json:"content"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func newReviewDTO(r *review.ReviewForQuery) ReviewDTO {
	dto := ReviewDTO{
		ID:        r.ID,
		CourseID:  r.CourseID,
		Grade:     r.Grade,
		Rating:    r.Rating,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if r.Course != nil {
		item := newCourseListItem(r.Course)
		dto.Course = &item
	}
	return dto
}

type CourseReviewFilter struct{}

type UserReviewFilter struct{}
