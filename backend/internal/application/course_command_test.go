package application_test

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/course"
)

type notifKey struct {
	userID, courseID int
}

type fakeNotificationRepo struct {
	levels map[notifKey]course.NotificationLevel
}

func newFakeNotificationRepo() *fakeNotificationRepo {
	return &fakeNotificationRepo{levels: make(map[notifKey]course.NotificationLevel)}
}

func (m *fakeNotificationRepo) GetLevel(ctx context.Context, userID, courseID int) (course.NotificationLevel, error) {
	if level, ok := m.levels[notifKey{userID, courseID}]; ok {
		return level, nil
	}
	return course.NotificationLevelNormal, nil
}

func (m *fakeNotificationRepo) SetLevel(ctx context.Context, userID, courseID int, level course.NotificationLevel) error {
	m.levels[notifKey{userID, courseID}] = level
	return nil
}

func (m *fakeNotificationRepo) GetCoursesByLevel(ctx context.Context, userID int, level course.NotificationLevel) ([]int, error) {
	var ids []int
	for k, v := range m.levels {
		if k.userID == userID && v == level {
			ids = append(ids, k.courseID)
		}
	}
	return ids, nil
}

type fakeCourseRepo struct {
	courses map[int]*course.Course
}

func newFakeCourseRepo() *fakeCourseRepo {
	return &fakeCourseRepo{courses: make(map[int]*course.Course)}
}

func (r *fakeCourseRepo) Get(ctx context.Context, courseID int) (*course.Course, error) {
	if c, ok := r.courses[courseID]; ok {
		return c, nil
	}
	return nil, nil
}

func (r *fakeCourseRepo) OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error) {
	_, ok := r.courses[courseID]
	return ok, nil
}

func TestCourseCommandService_SetNotificationLevel_Success(t *testing.T) {
	courseRepo := newFakeCourseRepo()
	courseRepo.courses[1] = &course.Course{ID: 1, Code: "CS101", Name: "数据结构"}
	notificationRepo := newFakeNotificationRepo()
	svc := application.NewCourseCommandService(courseRepo, notificationRepo)
	ctx := context.Background()

	err := svc.SetNotificationLevel(ctx, 100, 1, course.NotificationLevelFollow)
	if err != nil {
		t.Fatalf("SetNotificationLevel: %v", err)
	}

	level, _ := notificationRepo.GetLevel(ctx, 100, 1)
	if level != course.NotificationLevelFollow {
		t.Errorf("level: got %d, want %d", level, course.NotificationLevelFollow)
	}
}

func TestCourseCommandService_SetNotificationLevel_CourseNotFound(t *testing.T) {
	courseRepo := newFakeCourseRepo()
	notificationRepo := newFakeNotificationRepo()
	svc := application.NewCourseCommandService(courseRepo, notificationRepo)
	ctx := context.Background()

	err := svc.SetNotificationLevel(ctx, 100, 999, course.NotificationLevelFollow)
	if !errors.Is(err, application.ErrCourseNotFound) {
		t.Errorf("error: got %v, want %v", err, application.ErrCourseNotFound)
	}

	if len(notificationRepo.levels) != 0 {
		t.Error("expected no notification level set when course not found")
	}
}

func TestCourseCommandService_SetNotificationLevel_Update(t *testing.T) {
	courseRepo := newFakeCourseRepo()
	courseRepo.courses[1] = &course.Course{ID: 1}
	notificationRepo := newFakeNotificationRepo()
	notificationRepo.SetLevel(context.Background(), 100, 1, course.NotificationLevelFollow)
	svc := application.NewCourseCommandService(courseRepo, notificationRepo)
	ctx := context.Background()

	err := svc.SetNotificationLevel(ctx, 100, 1, course.NotificationLevelIgnored)
	if err != nil {
		t.Fatalf("SetNotificationLevel: %v", err)
	}

	level, _ := notificationRepo.GetLevel(ctx, 100, 1)
	if level != course.NotificationLevelIgnored {
		t.Errorf("level: got %d, want %d", level, course.NotificationLevelIgnored)
	}
}
