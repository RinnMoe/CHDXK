package application

import "jcourse/internal/domain/course"

type CreateCourseEnrollmentCommand struct {
	CourseID int    `json:"course_id"`
	Semester string `json:"semester"`
}

var (
	ErrSemesterRequired      = course.ErrSemesterRequired
	ErrOfferedCourseNotFound = course.ErrOfferedCourseNotFound
)
