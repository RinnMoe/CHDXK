package application

import "jcourse/internal/domain/course"

var (
	ErrSemesterRequired       = course.ErrSemesterRequired
	ErrInvalidSyncSemester    = course.ErrInvalidSyncSemester
	ErrEnrollmentSyncDisabled = course.ErrEnrollmentSyncDisabled
)
