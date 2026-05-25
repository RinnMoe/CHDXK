package application

import (
	"context"

	"jcourse/internal/domain/course"
)

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
