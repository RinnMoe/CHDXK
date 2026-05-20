package course

import "jcourse/internal/domain/teacher"

type OfferedCourse struct {
	ID         int
	CourseID   int
	Semester   string
	Language   string
	Grades     []string
	Categories []string
}

type OfferedCourseForQuery struct {
	ID           int
	Semester     string
	Language     string
	Grades       []string
	Categories   []string
	TeacherGroup []*teacher.TeacherForQuery
}
