package repository_test

import (
	"context"
	"testing"

	"jcourse/internal/domain/course"
	"jcourse/internal/infrastructure/repository"
)

func TestCourseNotificationRepository_GetLevel(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseNotificationRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "course_notifications")

	t.Run("returns normal when no record exists", func(t *testing.T) {
		level, err := repo.GetLevel(ctx, 1, 1)
		if err != nil {
			t.Fatalf("GetLevel: %v", err)
		}
		if level != course.NotificationLevelNormal {
			t.Errorf("level: got %d, want %d", level, course.NotificationLevelNormal)
		}
	})
}

func TestCourseNotificationRepository_SetLevel(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseNotificationRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "course_notifications")

	t.Run("creates new record", func(t *testing.T) {
		err := repo.SetLevel(ctx, 1, 1, course.NotificationLevelFollow)
		if err != nil {
			t.Fatalf("SetLevel: %v", err)
		}

		level, err := repo.GetLevel(ctx, 1, 1)
		if err != nil {
			t.Fatalf("GetLevel: %v", err)
		}
		if level != course.NotificationLevelFollow {
			t.Errorf("level: got %d, want %d", level, course.NotificationLevelFollow)
		}
	})

	t.Run("updates existing record", func(t *testing.T) {
		err := repo.SetLevel(ctx, 1, 1, course.NotificationLevelIgnored)
		if err != nil {
			t.Fatalf("SetLevel: %v", err)
		}

		level, err := repo.GetLevel(ctx, 1, 1)
		if err != nil {
			t.Fatalf("GetLevel: %v", err)
		}
		if level != course.NotificationLevelIgnored {
			t.Errorf("level: got %d, want %d", level, course.NotificationLevelIgnored)
		}
	})

	t.Run("resets to normal", func(t *testing.T) {
		err := repo.SetLevel(ctx, 1, 1, course.NotificationLevelNormal)
		if err != nil {
			t.Fatalf("SetLevel: %v", err)
		}

		level, err := repo.GetLevel(ctx, 1, 1)
		if err != nil {
			t.Fatalf("GetLevel: %v", err)
		}
		if level != course.NotificationLevelNormal {
			t.Errorf("level: got %d, want %d", level, course.NotificationLevelNormal)
		}
	})
}

func TestCourseNotificationRepository_GetCoursesByLevel(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseNotificationRepository(db)
	ctx := context.Background()

	cleanTables(t, db, "course_notifications")

	t.Run("returns empty when no records", func(t *testing.T) {
		ids, err := repo.GetCoursesByLevel(ctx, 1, course.NotificationLevelFollow)
		if err != nil {
			t.Fatalf("GetCoursesByLevel: %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("count: got %d, want 0", len(ids))
		}
	})

	t.Run("returns followed courses", func(t *testing.T) {
		repo.SetLevel(ctx, 1, 1, course.NotificationLevelFollow)
		repo.SetLevel(ctx, 1, 2, course.NotificationLevelIgnored)
		repo.SetLevel(ctx, 1, 3, course.NotificationLevelFollow)
		repo.SetLevel(ctx, 2, 4, course.NotificationLevelFollow)

		ids, err := repo.GetCoursesByLevel(ctx, 1, course.NotificationLevelFollow)
		if err != nil {
			t.Fatalf("GetCoursesByLevel: %v", err)
		}
		if len(ids) != 2 {
			t.Errorf("count: got %d, want 2", len(ids))
		}
	})

	t.Run("returns ignored courses", func(t *testing.T) {
		ids, err := repo.GetCoursesByLevel(ctx, 1, course.NotificationLevelIgnored)
		if err != nil {
			t.Fatalf("GetCoursesByLevel: %v", err)
		}
		if len(ids) != 1 {
			t.Errorf("count: got %d, want 1", len(ids))
		}
		if ids[0] != 2 {
			t.Errorf("course id: got %d, want 2", ids[0])
		}
	})

	t.Run("does not return normal courses", func(t *testing.T) {
		repo.SetLevel(ctx, 1, 5, course.NotificationLevelNormal)
		ids, err := repo.GetCoursesByLevel(ctx, 1, course.NotificationLevelNormal)
		if err != nil {
			t.Fatalf("GetCoursesByLevel: %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("count: got %d, want 0", len(ids))
		}
	})
}
