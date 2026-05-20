package course

import (
	"context"

	"jcourse/internal/domain/teacher"
)

type RatingInfo struct {
	Count        int
	Avg          float64
	Distribution [5]int
}

type CourseFilter struct {
	TeacherID   int
	ExcludeID   int
	Code        string
	Department  string
	Categories  []string
	Language    string
	TargetYears []string
	Credit      *float32
	HasReview   *bool
	OrderBy     string // "review_count" | "avg_rating"
	OrderDir    string // "asc" | "desc"
	Page        int
	PageSize    int
}

// Read model: course list/search result
type CourseView struct {
	ID            int
	Code          string
	Name          string
	Credit        float32
	Department    string
	MainTeacherID int
	MainTeacher   *teacher.TeacherView
	Categories    []string
	Language      string
	TargetYears   []string
	Rating        RatingInfo
}

// Read model: course detail with offered courses and rating distribution
type CourseDetailView struct {
	ID             int
	Code           string
	Name           string
	Credit         float32
	Department     string
	MainTeacherID  int
	MainTeacher    *teacher.TeacherView
	Categories     []string
	Language       string
	TargetYears    []string
	Rating         RatingInfo
	OfferedCourses []*OfferedCourseView
}

// Read model interface
type CourseQuery interface {
	FindBy(ctx context.Context, filter CourseFilter) ([]CourseView, int64, error)
	GetDetail(ctx context.Context, courseID int) (*CourseDetailView, error)
	FindOfferedCourses(ctx context.Context, courseID int) ([]OfferedCourseView, error)
}
