package application_test

import (
	"context"
	"slices"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/teacher"
)

type fakeCourseQuery struct {
	details map[int]*course.CourseDetailView
	views   map[int]course.CourseView
}

func newFakeCourseQuery() *fakeCourseQuery {
	return &fakeCourseQuery{
		details: make(map[int]*course.CourseDetailView),
		views:   make(map[int]course.CourseView),
	}
}

func (q *fakeCourseQuery) FindBy(ctx context.Context, filter course.CourseFilter) ([]course.CourseView, int64, error) {
	var results []course.CourseView
	for id, v := range q.views {
		if len(filter.CourseIDs) > 0 {
			match := slices.Contains(filter.CourseIDs, id)
			if !match {
				continue
			}
		}
		if filter.ExcludeID != 0 && filter.ExcludeID == id {
			continue
		}
		if filter.Code != "" && filter.Code != v.Code {
			continue
		}
		if filter.TeacherID != 0 && filter.TeacherID != v.MainTeacherID {
			continue
		}
		results = append(results, v)
	}
	return results, int64(len(results)), nil
}

func (q *fakeCourseQuery) GetDetail(ctx context.Context, courseID int) (*course.CourseDetailView, error) {
	if d, ok := q.details[courseID]; ok {
		dCopy := *d
		return &dCopy, nil
	}
	return nil, errNotFound
}

func (q *fakeCourseQuery) FindOfferedCourses(ctx context.Context, courseID int) ([]course.OfferedCourseView, error) {
	return nil, nil
}

var errNotFound = errorString("not found")

type errorString string

func (e errorString) Error() string { return string(e) }

func TestCourseQueryService_GetCourseDetail_WithUser(t *testing.T) {
	query := newFakeCourseQuery()
	query.details[1] = &course.CourseDetailView{
		ID:          1,
		Code:        "CS101",
		Name:        "数据结构",
		MainTeacher: &teacher.TeacherView{ID: 1, Name: "张三"},
	}
	notifRepo := newFakeNotificationRepo()
	notifRepo.SetLevel(context.Background(), 100, 1, course.NotificationLevelFollow)
	svc := application.NewCourseQueryService(query, notifRepo)
	ctx := context.Background()

	user := &auth.User{ID: 100}
	dto, err := svc.GetCourseDetail(ctx, user, 1)
	if err != nil {
		t.Fatalf("GetCourseDetail: %v", err)
	}
	if dto.NotificationLevel != int(course.NotificationLevelFollow) {
		t.Errorf("NotificationLevel: got %d, want %d", dto.NotificationLevel, course.NotificationLevelFollow)
	}
}

func TestCourseQueryService_GetCourseDetail_WithoutUser(t *testing.T) {
	query := newFakeCourseQuery()
	query.details[1] = &course.CourseDetailView{
		ID:          1,
		Code:        "CS101",
		Name:        "数据结构",
		MainTeacher: &teacher.TeacherView{ID: 1, Name: "张三"},
	}
	notifRepo := newFakeNotificationRepo()
	svc := application.NewCourseQueryService(query, notifRepo)
	ctx := context.Background()

	dto, err := svc.GetCourseDetail(ctx, nil, 1)
	if err != nil {
		t.Fatalf("GetCourseDetail: %v", err)
	}
	if dto.NotificationLevel != int(course.NotificationLevelNormal) {
		t.Errorf("NotificationLevel: got %d, want %d", dto.NotificationLevel, course.NotificationLevelNormal)
	}
}

func TestCourseQueryService_ListCoursesByNotificationLevel(t *testing.T) {
	query := newFakeCourseQuery()
	query.views[1] = course.CourseView{ID: 1, Code: "CS101", Name: "数据结构"}
	query.views[2] = course.CourseView{ID: 2, Code: "CS102", Name: "算法"}
	query.views[3] = course.CourseView{ID: 3, Code: "CS103", Name: "操作系统"}

	notifRepo := newFakeNotificationRepo()
	ctx := context.Background()
	notifRepo.SetLevel(ctx, 100, 1, course.NotificationLevelFollow)
	notifRepo.SetLevel(ctx, 100, 2, course.NotificationLevelIgnored)
	notifRepo.SetLevel(ctx, 100, 3, course.NotificationLevelFollow)

	svc := application.NewCourseQueryService(query, notifRepo)

	t.Run("list followed", func(t *testing.T) {
		result, err := svc.ListCoursesByNotificationLevel(ctx, 100, course.NotificationLevelFollow, application.CourseListFilter{})
		if err != nil {
			t.Fatalf("ListCoursesByNotificationLevel: %v", err)
		}
		if result.Total != 2 {
			t.Errorf("total: got %d, want 2", result.Total)
		}
	})

	t.Run("list ignored", func(t *testing.T) {
		result, err := svc.ListCoursesByNotificationLevel(ctx, 100, course.NotificationLevelIgnored, application.CourseListFilter{})
		if err != nil {
			t.Fatalf("ListCoursesByNotificationLevel: %v", err)
		}
		if result.Total != 1 {
			t.Errorf("total: got %d, want 1", result.Total)
		}
	})

	t.Run("returns empty when no levels set", func(t *testing.T) {
		result, err := svc.ListCoursesByNotificationLevel(ctx, 999, course.NotificationLevelFollow, application.CourseListFilter{})
		if err != nil {
			t.Fatalf("ListCoursesByNotificationLevel: %v", err)
		}
		if result.Total != 0 {
			t.Errorf("total: got %d, want 0", result.Total)
		}
		if result.Items == nil {
			t.Error("Items should not be nil")
		}
	})
}
