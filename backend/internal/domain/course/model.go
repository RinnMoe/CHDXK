package course

import (
	"context"
)

type CourseRepository interface {
	Get(ctx context.Context, courseID int) (*CourseView, error)
	FindBy(ctx context.Context, filter CourseFilter) ([]CourseView, int64, error)
	GetDetail(ctx context.Context, courseID int) (*CourseDetailView, error)
	FindOfferedCourses(ctx context.Context, courseID int) ([]OfferedCourseView, error)
	GetFilters(ctx context.Context) (*CourseFilters, error)
	UpdateModeratorRemark(ctx context.Context, courseID int, moderatorRemark string) error
	RefreshRatingScores(ctx context.Context, config RatingScoreConfig) error
	OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error)
	OfferedSemesterExists(ctx context.Context, semester string) (bool, error)
}

type CourseCodeTeacher struct {
	Code        string
	TeacherName string
}
