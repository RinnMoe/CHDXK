package application_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/teacher"
)

type fakeCourseQuery struct {
	details map[int]*course.CourseDetailView
	views   map[int]course.CourseView
}

type fakeHotCourseRepo struct {
	ranks []course.HotCourseRank
}

func (r *fakeHotCourseRepo) AddScore(ctx context.Context, courseID int, score int64, at time.Time) error {
	return nil
}

func (r *fakeHotCourseRepo) Top(ctx context.Context, period course.HotCoursePeriod, at time.Time, limit int64) ([]course.HotCourseRank, error) {
	if limit < int64(len(r.ranks)) {
		return r.ranks[:limit], nil
	}
	return r.ranks, nil
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

func (q *fakeCourseQuery) GetFilters(ctx context.Context) (*course.CourseFilters, error) {
	return &course.CourseFilters{}, nil
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
	svc := application.NewCourseQueryService(query, notifRepo, nil)
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
	svc := application.NewCourseQueryService(query, notifRepo, nil)
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

	svc := application.NewCourseQueryService(query, notifRepo, nil)

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

func TestCourseQueryService_ListHotCourses(t *testing.T) {
	query := newFakeCourseQuery()
	query.views[1] = course.CourseView{ID: 1, Code: "CS101", Name: "数据结构"}
	query.views[2] = course.CourseView{ID: 2, Code: "CS102", Name: "算法"}

	hotRepo := &fakeHotCourseRepo{ranks: []course.HotCourseRank{
		{CourseID: 2, Score: 10},
		{CourseID: 1, Score: 8},
	}}
	svc := application.NewCourseQueryService(query, newFakeNotificationRepo(), hotRepo)

	result, err := svc.ListHotCourses(context.Background(), "week")
	if err != nil {
		t.Fatalf("ListHotCourses: %v", err)
	}
	if result.Period != "week" {
		t.Fatalf("Period: got %s, want week", result.Period)
	}
	if len(result.Items) != 2 {
		t.Fatalf("items length: got %d, want 2", len(result.Items))
	}
	if result.Items[0].Course.ID != 2 || result.Items[0].Score != 10 {
		t.Fatalf("first item: got course=%d score=%d, want course=2 score=10", result.Items[0].Course.ID, result.Items[0].Score)
	}
	if result.Items[1].Course.ID != 1 || result.Items[1].Score != 8 {
		t.Fatalf("second item: got course=%d score=%d, want course=1 score=8", result.Items[1].Course.ID, result.Items[1].Score)
	}
}

func TestCourseQueryService_ListHotCourses_InvalidPeriod(t *testing.T) {
	svc := application.NewCourseQueryService(newFakeCourseQuery(), newFakeNotificationRepo(), &fakeHotCourseRepo{})
	_, err := svc.ListHotCourses(context.Background(), "daily")
	if err != course.ErrInvalidHotCoursePeriod {
		t.Fatalf("err: got %v, want %v", err, course.ErrInvalidHotCoursePeriod)
	}
}
