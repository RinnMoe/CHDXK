package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/course"
)

type CourseNotificationRepository struct {
	db *gorm.DB
}

func NewCourseNotificationRepository(db *gorm.DB) *CourseNotificationRepository {
	return &CourseNotificationRepository{db: db}
}

func (r *CourseNotificationRepository) GetLevel(ctx context.Context, userID, courseID int) (course.NotificationLevel, error) {
	e, err := gorm.G[CourseNotificationEntity](r.db).Where("user_id = ? AND course_id = ?", userID, courseID).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return course.NotificationLevelNormal, nil
		}
		return 0, err
	}
	return course.NotificationLevel(e.Level), nil
}

func (r *CourseNotificationRepository) SetLevel(ctx context.Context, userID, courseID int, level course.NotificationLevel) error {
	if level == course.NotificationLevelNormal {
		_, err := gorm.G[CourseNotificationEntity](r.db).
			Where("user_id = ? AND course_id = ?", userID, courseID).
			Delete(ctx)
		return err
	}
	e := CourseNotificationEntity{
		UserID:    userID,
		CourseID:  courseID,
		Level:     int(level),
		UpdatedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Save(&e).Error
}

func (r *CourseNotificationRepository) GetCoursesByLevel(ctx context.Context, userID int, level course.NotificationLevel) ([]int, error) {
	entities, err := gorm.G[CourseNotificationEntity](r.db).
		Where("user_id = ? AND level = ?", userID, int(level)).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]int, len(entities))
	for i, e := range entities {
		ids[i] = e.CourseID
	}
	return ids, nil
}

var _ course.CourseNotificationRepository = (*CourseNotificationRepository)(nil)
