package application

import (
	"context"
	"time"

	"jcourse/internal/domain/course"
)

type CourseEnrollmentCommandService struct {
	enrollments *course.EnrollmentService
}

func NewCourseEnrollmentCommandService(courseRepo course.CourseRepository, enrollmentRepo course.CourseEnrollmentRepository) *CourseEnrollmentCommandService {
	return &CourseEnrollmentCommandService{enrollments: course.NewEnrollmentService(courseRepo, enrollmentRepo)}
}

func (s *CourseEnrollmentCommandService) Create(ctx context.Context, userID int, cmd CreateCourseEnrollmentCommand) error {
	return s.enrollments.Create(ctx, userID, cmd.CourseID, cmd.Semester, time.Now())
}

func (s *CourseEnrollmentCommandService) Delete(ctx context.Context, userID, enrollmentID int) error {
	return s.enrollments.Delete(ctx, userID, enrollmentID)
}
