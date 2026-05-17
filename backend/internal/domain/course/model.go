package course

import "context"

type Course struct {
	ID            int
	Code          string
	Name          string
	Credit        float32
	MainTeacherID int
	CreatedAt     int64
}

type CourseRepository interface {
	Get(ctx context.Context, courseID int) (*Course, error)
}
