package course

import (
	"context"
	"time"
)

type CourseEnrollment struct {
	ID        int
	UserID    int
	CourseID  int
	Semester  string
	CreatedAt time.Time
}

type CourseEnrollmentView struct {
	ID        int
	UserID    int
	Course    CourseView
	Semester  string
	CreatedAt time.Time
}

type CourseEnrollmentFilter struct {
	UserID   int
	CourseID int
}

type CourseEnrollmentRepository interface {
	Create(ctx context.Context, enrollment *CourseEnrollment) error
	Delete(ctx context.Context, enrollmentID, userID int) error
}

type CourseEnrollmentQuery interface {
	FindUserEnrollments(ctx context.Context, userID int) ([]CourseEnrollmentView, error)
	FindUserCourseEnrollments(ctx context.Context, userID, courseID int) ([]CourseEnrollmentView, error)
}
