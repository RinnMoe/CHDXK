package review

import (
	"context"
	"time"

	"jcourse/internal/domain/course"
)

type ReviewQuery interface {
	FindBy(ctx context.Context, filter ReviewFilter) ([]ReviewView, int64, error)
	GetByID(ctx context.Context, reviewID int) (*ReviewView, error)
	GetCourseFilters(ctx context.Context, courseID int) (*ReviewFilters, error)
	FindRevisions(ctx context.Context, reviewID int) ([]RevisionView, error)
}

type FilterItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type ReviewFilters struct {
	Semesters []FilterItem `json:"semesters"`
	Ratings   []FilterItem `json:"ratings"`
}

type ReviewFilter struct {
	ReviewID         int
	CourseID         int
	CourseIDs        []int
	ExcludeCourseIDs []int
	UserID           int
	Q                string
	Semester         string
	Rating           int
	CreatedAfter     time.Time
	OrderBy          string
	Ascend           bool
	Page             int
	PageSize         int
	WithCourse       bool
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
