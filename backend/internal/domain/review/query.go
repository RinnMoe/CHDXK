package review

import (
	"context"
	"time"

	"jcourse/internal/domain/course"
)

type ReviewQuery interface {
	FindBy(ctx context.Context, filter ReviewFilter) ([]ReviewForQuery, error)
	FindRevisions(ctx context.Context, reviewID int) ([]RevisionForQuery, error)
}

type ReviewFilter struct {
	ReviewID int
	CourseID int
	UserID   int
	Semester string
	Rating   int
	Order    string
}

type ReviewForQuery struct {
	Review
	Course *course.CourseForQuery
}

type RevisionForQuery struct {
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
