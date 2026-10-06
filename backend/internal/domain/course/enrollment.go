package course

import (
	"context"
	"time"

	"jcourse/pkg/apperr"
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
	Save(ctx context.Context, enrollmentID, userID, sectionID int) error
	Delete(ctx context.Context, enrollmentID, userID int) error
}

func (s *EnrollmentService) Save(ctx context.Context, userID, enrollmentID, sectionID int) error {
	if userID <= 0 || enrollmentID < 0 || sectionID <= 0 {
		return apperr.BadRequest("请选择系统已有的课程和教学班")
	}
	return s.enrollmentRepo.Save(ctx, enrollmentID, userID, sectionID)
}

type CourseEnrollmentQuery interface {
	FindUserEnrollments(ctx context.Context, userID int) ([]CourseEnrollmentView, error)
	FindUserCourseEnrollments(ctx context.Context, userID, courseID int) ([]CourseEnrollmentView, error)
}

type EnrollmentService struct {
	enrollmentRepo CourseEnrollmentRepository
}

func NewEnrollmentService(enrollmentRepo CourseEnrollmentRepository) *EnrollmentService {
	return &EnrollmentService{enrollmentRepo: enrollmentRepo}
}

func (s *EnrollmentService) Delete(ctx context.Context, userID, enrollmentID int) error {
	return s.enrollmentRepo.Delete(ctx, enrollmentID, userID)
}
