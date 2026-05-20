package course

import (
	"context"
	"time"
)

type Course struct {
	ID            int
	Code          string
	Name          string
	Credit        float32
	MainTeacherID int
	Categories    []string
	Language      string
	Grades        []string
	RatingCount   int
	RatingAvg     float64
	CreatedAt     time.Time
}

type CourseRepository interface {
	Get(ctx context.Context, courseID int) (*Course, error)
	OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error)
}
