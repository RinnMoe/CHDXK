package course

import "jcourse/internal/domain/teacher"

// Write model: offered course aggregate
type OfferedCourse struct {
	ID         int
	CourseID   int
	Semester   string
	Language   string
	Grades     []string
	Categories []string
}

// Read model: offered course query result with teacher group
type OfferedCourseView struct {
	ID           int
	Semester     string
	Language     string
	Grades       []string
	Categories   []string
	TeacherGroup []*teacher.TeacherView
}
