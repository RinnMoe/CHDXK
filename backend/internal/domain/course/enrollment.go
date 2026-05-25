package course

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrSemesterRequired      = errors.New("semester required")
	ErrOfferedCourseNotFound = errors.New("offered course not found")
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
	SyncFromCoursePairs(ctx context.Context, userID int, semester string, pairs []CourseCodeTeacher) (int64, error)
	Delete(ctx context.Context, enrollmentID, userID int) error
}

type CourseEnrollmentQuery interface {
	FindUserEnrollments(ctx context.Context, userID int) ([]CourseEnrollmentView, error)
	FindUserCourseEnrollments(ctx context.Context, userID, courseID int) ([]CourseEnrollmentView, error)
}

type EnrollmentService struct {
	courseRepo     CourseRepository
	enrollmentRepo CourseEnrollmentRepository
}

func NewEnrollmentService(courseRepo CourseRepository, enrollmentRepo CourseEnrollmentRepository) *EnrollmentService {
	return &EnrollmentService{courseRepo: courseRepo, enrollmentRepo: enrollmentRepo}
}

func (s *EnrollmentService) Create(ctx context.Context, userID, courseID int, semester string, now time.Time) error {
	semester = strings.TrimSpace(semester)
	if semester == "" {
		return ErrSemesterRequired
	}
	c, err := s.courseRepo.Get(ctx, courseID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCourseNotFound
	}
	exists := semester == c.LastSemester
	if !exists {
		exists, err = s.courseRepo.OfferedCourseExists(ctx, courseID, semester)
	}
	if err != nil {
		return err
	}
	if !exists {
		return ErrOfferedCourseNotFound
	}
	return s.enrollmentRepo.Create(ctx, &CourseEnrollment{
		UserID:    userID,
		CourseID:  courseID,
		Semester:  semester,
		CreatedAt: now,
	})
}

func (s *EnrollmentService) Delete(ctx context.Context, userID, enrollmentID int) error {
	return s.enrollmentRepo.Delete(ctx, enrollmentID, userID)
}
