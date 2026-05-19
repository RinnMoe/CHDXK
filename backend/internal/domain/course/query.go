package course

import (
	"context"

	"jcourse/internal/domain/teacher"
)

type CourseFilter struct {
	TeacherID  int
	Code       string
	Department string
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
	ReviewCount   int
	AvgRating     float64
}

type CourseDetailForQuery struct {
	CourseForQuery
	RatingDistribution [5]int // index 0 = 1-star, ..., index 4 = 5-star
	OtherTeachers      []TeacherCourseStats
	OtherCourses       []CourseForQuery
}

type TeacherCourseStats struct {
	Teacher     teacher.TeacherForQuery
	CourseID    int
	ReviewCount int
	AvgRating   float64
}

type CourseQuery interface {
	FindBy(ctx context.Context, filter CourseFilter) ([]CourseForQuery, int64, error)
	GetDetail(ctx context.Context, courseID int) (*CourseDetailForQuery, error)
}
