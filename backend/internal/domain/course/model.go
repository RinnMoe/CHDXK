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
	CreatedAt     time.Time
}

type CourseRepository interface {
	Get(ctx context.Context, courseID int) (*Course, error)
}
