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
	ReviewID int
	CourseID int
	UserID   int
	Semester string
	Rating   int
	Order    string
}

// Read model: review with eager-loaded course summary
type ReviewView struct {
	Review
	Course *course.CourseView
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
	Grade     string
	CreatedAt time.Time
}
