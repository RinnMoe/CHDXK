package course

import "context"

type CourseFilter struct{}

type CourseForQuery struct {
	ID            int
	Code          string
	Name          string
	Credit        float32
	MainTeacherID int
	MainTeacher   *TeacherForQuery
}

type CourseQuery interface {
	FindBy(ctx context.Context, filter CourseFilter) ([]CourseForQuery, error)
}

type TeacherForQuery struct {
	ID         int
	Code       string
	Name       string
	Department string
}
