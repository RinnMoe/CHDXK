package course

import (
	"context"
	"errors"
)

var ErrCourseNotFound = errors.New("course not found")

type NotificationLevel int

const (
	NotificationLevelNormal  NotificationLevel = 0
	NotificationLevelFollow  NotificationLevel = 1
	NotificationLevelIgnored NotificationLevel = 2
)

type CourseNotification struct {
	UserID   int
	CourseID int
	Level    NotificationLevel
}

type CourseNotificationRepository interface {
	GetLevel(ctx context.Context, userID, courseID int) (NotificationLevel, error)
	SetLevel(ctx context.Context, userID, courseID int, level NotificationLevel) error
	GetCoursesByLevel(ctx context.Context, userID int, level NotificationLevel) ([]int, error)
}

type NotificationService struct {
	courseRepo       CourseRepository
	notificationRepo CourseNotificationRepository
}

func NewNotificationService(courseRepo CourseRepository, notificationRepo CourseNotificationRepository) *NotificationService {
	return &NotificationService{courseRepo: courseRepo, notificationRepo: notificationRepo}
}

func (s *NotificationService) SetLevel(ctx context.Context, userID, courseID int, level NotificationLevel) error {
	c, err := s.courseRepo.Get(ctx, courseID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCourseNotFound
	}
	return s.notificationRepo.SetLevel(ctx, userID, courseID, level)
}
