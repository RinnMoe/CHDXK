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
	TeacherID  int
	ExcludeID  int
	Code       string
	Department string
	Categories []string
	Language   string
	Grades     []string
	Credit     *float32
	HasReview  *bool
	OrderBy    string // "review_count" | "avg_rating"
	OrderDir   string // "asc" | "desc"
	Page       int
	PageSize   int
}

type CourseForQuery struct {
	ID            int
	Code          string
	Name          string
	Credit        float32
	Department    string
	MainTeacherID int
	MainTeacher   *teacher.TeacherForQuery
	Categories    []string
	Language      string
	Grades        []string
	Rating        RatingInfo
}

type CourseDetailForQuery struct {
	ID             int
	Code           string
	Name           string
	Credit         float32
	Department     string
	MainTeacherID  int
	MainTeacher    *teacher.TeacherForQuery
	Categories     []string
	Language       string
	Grades         []string
	Rating         RatingInfo
	OfferedCourses []*OfferedCourseForQuery
}

type CourseQuery interface {
	FindBy(ctx context.Context, filter CourseFilter) ([]CourseForQuery, int64, error)
	GetDetail(ctx context.Context, courseID int) (*CourseDetailForQuery, error)
	FindOfferedCourses(ctx context.Context, courseID int) ([]OfferedCourseForQuery, error)
}
