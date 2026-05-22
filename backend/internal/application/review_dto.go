package application

import (
	"time"

	"jcourse/internal/domain/review"
)

type VoteStats struct {
	LikeCount    int  `json:"like_count"`
	DislikeCount int  `json:"dislike_count"`
	MyVote       *int `json:"my_vote,omitempty"`
}

type ReviewDTO struct {
	ID        int                `json:"id"`
	Course    *CourseListItemDTO `json:"course,omitempty"`
	CourseID  int                `json:"course_id"`
	UserID    int                `json:"user_id,omitempty"`
	Semester  string             `json:"semester"`
	Score     string             `json:"score"`
	Rating    int                `json:"rating"`
	Content   string             `json:"content"`
	Vote      VoteStats          `json:"vote"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func newReviewDTO(r *review.ReviewView, showUserID bool) ReviewDTO {
	dto := ReviewDTO{
		ID:       r.ID,
		CourseID: r.CourseID,
		Semester: r.Semester,
		Score:    r.Score,
		Rating:   r.Rating,
		Content:  r.Content,
		Vote: VoteStats{
			LikeCount:    r.Vote.LikeCount,
			DislikeCount: r.Vote.DislikeCount,
		},
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if showUserID {
		dto.UserID = r.UserID
	}
	if r.Course != nil {
		item := newCourseListItemDTO(r.Course)
		dto.Course = &item
	}
	return dto
}
