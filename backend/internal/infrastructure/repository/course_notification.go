package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"jcourse/internal/domain/course"
)

type CourseNotificationRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewCourseNotificationRepository(db *gorm.DB, cache ...*redis.Client) *CourseNotificationRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &CourseNotificationRepository{db: db, cache: client}
}

func (r *CourseNotificationRepository) GetLevel(ctx context.Context, userID, courseID int) (course.NotificationLevel, error) {
	key := cacheKey("course_notification", userID, courseID, "level")
	if cached, ok := cacheGetJSON[course.NotificationLevel](ctx, r.cache, key); ok {
		return *cached, nil
	}

	e, err := gorm.G[CourseNotificationEntity](r.db).Where("user_id = ? AND course_id = ?", userID, courseID).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cacheSetJSON(ctx, r.cache, key, course.NotificationLevelNormal)
			return course.NotificationLevelNormal, nil
		}
		return 0, err
	}
	level := course.NotificationLevel(e.Level)
	cacheSetJSON(ctx, r.cache, key, level)
	return level, nil
}

func (r *CourseNotificationRepository) SetLevel(ctx context.Context, userID, courseID int, level course.NotificationLevel) error {
	deleteCache := func() {
		cacheDelete(ctx, r.cache, cacheKey("course_notification", userID, courseID, "level"))
		cacheDeletePattern(ctx, r.cache, cacheKey("course_notification", userID, "level", "*")+":courses")
	}
	if level == course.NotificationLevelNormal {
		_, err := gorm.G[CourseNotificationEntity](r.db).
			Where("user_id = ? AND course_id = ?", userID, courseID).
			Delete(ctx)
		if err == nil {
			deleteCache()
		}
		return err
	}
	e := CourseNotificationEntity{
		UserID:    userID,
		CourseID:  courseID,
		Level:     int(level),
		UpdatedAt: time.Now(),
	}
	if err := r.db.WithContext(ctx).Save(&e).Error; err != nil {
		return err
	}
	deleteCache()
	return nil
}

func (r *CourseNotificationRepository) GetCoursesByLevel(ctx context.Context, userID int, level course.NotificationLevel) ([]int, error) {
	key := cacheKey("course_notification", userID, "level", level, "courses")
	if cached, ok := cacheGetJSON[[]int](ctx, r.cache, key); ok {
		return *cached, nil
	}

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
	cacheSetJSON(ctx, r.cache, key, ids)
	return ids, nil
}

var _ course.CourseNotificationRepository = (*CourseNotificationRepository)(nil)
