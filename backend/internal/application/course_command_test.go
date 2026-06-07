package application_test

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
)

func newFakeNotificationRepo() *course.MockCourseNotificationRepository {
	return course.NewMockCourseNotificationRepository()
}

func newFakeCourseRepo() *course.MockCourseRepository {
	repo := course.NewMockCourseRepository()
	repo.OnOfferedCourseExists = func(_ context.Context, courseID int, _ string) (bool, error) {
		_, ok := repo.Courses[courseID]
		return ok, nil
	}
	return repo
}

func TestCourseCommandService_SetNotificationLevel_Success(t *testing.T) {
	courseRepo := newFakeCourseRepo()
	courseRepo.Courses[1] = &course.CourseView{ID: 1, Code: "CS101", Name: "数据结构"}
	notificationRepo := newFakeNotificationRepo()
	svc := application.NewCourseCommandService(course.NewService(courseRepo), course.NewNotificationService(courseRepo, notificationRepo), nil, nil, nil, nil)
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
	svc := application.NewCourseCommandService(course.NewService(courseRepo), course.NewNotificationService(courseRepo, notificationRepo), nil, nil, nil, nil)
	ctx := context.Background()

	err := svc.SetNotificationLevel(ctx, 100, 999, course.NotificationLevelFollow)
	if !errors.Is(err, application.ErrCourseNotFound) {
		t.Errorf("error: got %v, want %v", err, application.ErrCourseNotFound)
	}

	if len(notificationRepo.Levels) != 0 {
		t.Error("expected no notification level set when course not found")
	}
}

func TestCourseCommandService_SetNotificationLevel_Update(t *testing.T) {
	courseRepo := newFakeCourseRepo()
	courseRepo.Courses[1] = &course.CourseView{ID: 1}
	notificationRepo := newFakeNotificationRepo()
	notificationRepo.SetLevel(context.Background(), 100, 1, course.NotificationLevelFollow)
	svc := application.NewCourseCommandService(course.NewService(courseRepo), course.NewNotificationService(courseRepo, notificationRepo), nil, nil, nil, nil)
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

func TestCourseCommandService_UpdateModeratorRemarkRequiresAdmin(t *testing.T) {
	courseRepo := newFakeCourseRepo()
	courseRepo.Courses[1] = &course.CourseView{ID: 1}
	notificationRepo := newFakeNotificationRepo()
	svc := application.NewCourseCommandService(course.NewService(courseRepo), course.NewNotificationService(courseRepo, notificationRepo), nil, nil, nil, nil)
	ctx := context.Background()

	err := svc.UpdateModeratorRemark(ctx, &auth.User{ID: 10, Role: auth.RoleUser}, 1, &application.UpdateCourseModeratorRemarkCommand{
		ModeratorRemark: "需要核实",
	})
	if !errors.Is(err, course.ErrUserCannotModerate) {
		t.Fatalf("non-admin UpdateModeratorRemark error = %v, want %v", err, course.ErrUserCannotModerate)
	}

	err = svc.UpdateModeratorRemark(ctx, &auth.User{ID: 99, Role: auth.RoleAdmin}, 1, &application.UpdateCourseModeratorRemarkCommand{
		ModeratorRemark: "已核实",
	})
	if err != nil {
		t.Fatalf("admin UpdateModeratorRemark: %v", err)
	}
	if got := courseRepo.Courses[1].ModeratorRemark; got != "已核实" {
		t.Fatalf("ModeratorRemark = %q, want 已核实", got)
	}
}
