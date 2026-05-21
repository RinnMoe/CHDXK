package course

import "context"

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
