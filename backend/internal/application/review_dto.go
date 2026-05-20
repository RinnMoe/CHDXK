package application

import (
	"time"

	"jcourse/internal/domain/review"
)

// Read model: review with eager-loaded course summary
type ReviewDTO struct {
	ID        int                `json:"id"`
	Course    *CourseListItemDTO `json:"course,omitempty"`
	CourseID  int                `json:"course_id"`
	Score     string             `json:"score"`
	Rating    int                `json:"rating"`
	Content   string             `json:"content"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func newReviewDTO(r *review.ReviewView) ReviewDTO {
	dto := ReviewDTO{
		ID:        r.ID,
		CourseID:  r.CourseID,
		Score:     r.Score,
		Rating:    r.Rating,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if r.Course != nil {
		item := newCourseListItemDTO(r.Course)
		dto.Course = &item
	}
	return dto
}
