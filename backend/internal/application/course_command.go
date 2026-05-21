package application

import (
	"context"
	"errors"

	"jcourse/internal/domain/course"
)

var ErrCourseNotFound = errors.New("course not found")

type CourseCommandService struct {
	courseRepo       course.CourseRepository
	notificationRepo course.CourseNotificationRepository
}

func NewCourseCommandService(courseRepo course.CourseRepository, notificationRepo course.CourseNotificationRepository) *CourseCommandService {
	return &CourseCommandService{
		courseRepo:       courseRepo,
		notificationRepo: notificationRepo,
	}
}

func (s *CourseCommandService) SetNotificationLevel(ctx context.Context, userID, courseID int, level course.NotificationLevel) error {
	_, err := s.courseRepo.Get(ctx, courseID)
	if err != nil {
		return ErrCourseNotFound
	}
	return s.notificationRepo.SetLevel(ctx, userID, courseID, level)
}
