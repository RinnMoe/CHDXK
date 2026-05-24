package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"jcourse/internal/domain/course"
)

type CreateCourseEnrollmentCommand struct {
	CourseID int    `json:"course_id"`
	Semester string `json:"semester"`
}

type CourseEnrollmentQueryService struct {
	enrollmentQuery course.CourseEnrollmentQuery
}

func NewCourseEnrollmentQueryService(enrollmentQuery course.CourseEnrollmentQuery) *CourseEnrollmentQueryService {
	return &CourseEnrollmentQueryService{enrollmentQuery: enrollmentQuery}
}

func (s *CourseEnrollmentQueryService) ListMyEnrollments(ctx context.Context, userID int) ([]CourseEnrollmentDTO, error) {
	rows, err := s.enrollmentQuery.FindUserEnrollments(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]CourseEnrollmentDTO, len(rows))
	for i := range rows {
		items[i] = newCourseEnrollmentDTO(&rows[i])
	}

	return items, nil
}

type CourseEnrollmentCommandService struct {
	courseRepo     course.CourseRepository
	enrollmentRepo course.CourseEnrollmentRepository
}

func NewCourseEnrollmentCommandService(courseRepo course.CourseRepository, enrollmentRepo course.CourseEnrollmentRepository) *CourseEnrollmentCommandService {
	return &CourseEnrollmentCommandService{courseRepo: courseRepo, enrollmentRepo: enrollmentRepo}
}

func (s *CourseEnrollmentCommandService) Create(ctx context.Context, userID int, cmd CreateCourseEnrollmentCommand) error {
	semester := strings.TrimSpace(cmd.Semester)
	if semester == "" {
		return ErrSemesterRequired
	}
	c, err := s.courseRepo.Get(ctx, cmd.CourseID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCourseNotFound
	}
	exists := semester == c.LastSemester
	if !exists {
		exists, err = s.courseRepo.OfferedCourseExists(ctx, cmd.CourseID, semester)
	}
	if err != nil {
		return err
	}
	if !exists {
		return ErrOfferedCourseNotFound
	}
	return s.enrollmentRepo.Create(ctx, &course.CourseEnrollment{
		UserID:    userID,
		CourseID:  cmd.CourseID,
		Semester:  semester,
		CreatedAt: time.Now(),
	})
}

func (s *CourseEnrollmentCommandService) Delete(ctx context.Context, userID, enrollmentID int) error {
	return s.enrollmentRepo.Delete(ctx, enrollmentID, userID)
}

var (
	ErrSemesterRequired      = errors.New("semester required")
	ErrOfferedCourseNotFound = errors.New("offered course not found")
)
