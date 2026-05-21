package application

import (
	"context"

	"jcourse/internal/domain/course"
)

var ErrCourseNotFound = course.ErrCourseNotFound

type CourseCommandService struct {
	notificationService *course.NotificationService
}

func NewCourseCommandService(courseRepo course.CourseRepository, notificationRepo course.CourseNotificationRepository) *CourseCommandService {
	return &CourseCommandService{
		notificationService: course.NewNotificationService(courseRepo, notificationRepo),
	}
}

func (s *CourseCommandService) SetNotificationLevel(ctx context.Context, userID, courseID int, level course.NotificationLevel) error {
	return s.notificationService.SetLevel(ctx, userID, courseID, level)
}
