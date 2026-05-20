package review

import (
	"context"
	"time"

	"jcourse/internal/domain/course"
)

type ReviewQuery interface {
	FindBy(ctx context.Context, filter ReviewFilter) ([]ReviewView, error)
	FindRevisions(ctx context.Context, reviewID int) ([]RevisionView, error)
}

type ReviewFilter struct {
	ReviewID     int
	CourseID     int
	UserID       int
	Semester     string
	Rating       int
	CreatedAfter time.Time
	Order        string
	Limit        int
}

// Read model: review with eager-loaded course summary
type ReviewView struct {
	ID        int
	CourseID  int
	Semester  string
	UserID    int
	Rating    int
	Content   string
	Score     string
	Vote      ReviewVoteStats
	CreatedAt time.Time
	UpdatedAt time.Time
	Course    *course.CourseView
}

type ReviewVoteStats struct {
	LikeCount    int
	DislikeCount int
}

// Read model: review revision snapshot
type RevisionView struct {
	ID        int
	ReviewID  int
	CourseID  int
	Semester  string
	UserID    int
	Rating    int
	Content   string
	Score     string
	CreatedAt time.Time
}
