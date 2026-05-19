package course

import "jcourse/internal/domain/teacher"

type OfferedCourse struct {
	ID       int
	CourseID int
	Semester string
	Language string
	Grade    string
}

type OfferedCourseForQuery struct {
	ID           int
	Semester     string
	Language     string
	Grade        string
	TeacherGroup []*teacher.TeacherForQuery
}
