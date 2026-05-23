package application_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/teacher"
)

type fakeCourseQuery struct {
	details map[int]*course.CourseDetailView
	views   map[int]course.CourseView
}

type fakeHotCourseRepo struct {
	ranks []course.HotCourseRank
}

type fakeTeacherQuery struct {
	views map[int]teacher.TeacherView
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

func newFakeTeacherQuery() *fakeTeacherQuery {
	return &fakeTeacherQuery{views: make(map[int]teacher.TeacherView)}
}

func (q *fakeTeacherQuery) FindBy(ctx context.Context, filter teacher.TeacherFilter) ([]teacher.TeacherView, int64, error) {
	var results []teacher.TeacherView
	for id, v := range q.views {
		if len(filter.TeacherIDs) > 0 && !slices.Contains(filter.TeacherIDs, id) {
			continue
		}
		results = append(results, v)
	}
	return results, int64(len(results)), nil
}

func (q *fakeTeacherQuery) GetFilters(ctx context.Context) (*teacher.TeacherFilters, error) {
	return &teacher.TeacherFilters{}, nil
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
	svc := application.NewCourseQueryService(query, nil, newFakeReviewQuery(), notifRepo, nil)
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

func TestCourseQueryService_GetCourseDetail_WithMyReview(t *testing.T) {
	query := newFakeCourseQuery()
	query.details[1] = &course.CourseDetailView{
		ID:          1,
		Code:        "CS101",
		Name:        "数据结构",
		MainTeacher: &teacher.TeacherView{ID: 1, Name: "张三"},
	}
	reviewQuery := newFakeReviewQuery()
	reviewQuery.reviews = []review.ReviewView{
		{ID: 10, CourseID: 1, UserID: 100, Rating: 5, Content: "很好"},
		{ID: 11, CourseID: 1, UserID: 101, Rating: 3, Content: "一般"},
		{ID: 12, CourseID: 2, UserID: 100, Rating: 4, Content: "还行"},
	}
	svc := application.NewCourseQueryService(query, nil, reviewQuery, newFakeNotificationRepo(), nil)

	dto, err := svc.GetCourseDetail(context.Background(), &auth.User{ID: 100}, 1)
	if err != nil {
		t.Fatalf("GetCourseDetail: %v", err)
	}
	if dto.MyReview == nil {
		t.Fatal("MyReview should be set")
	}
	if dto.MyReview.ID != 10 {
		t.Errorf("MyReview.ID: got %d, want 10", dto.MyReview.ID)
	}
	if dto.MyReview.UserID != 100 {
		t.Errorf("MyReview.UserID: got %d, want 100", dto.MyReview.UserID)
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
	svc := application.NewCourseQueryService(query, nil, newFakeReviewQuery(), notifRepo, nil)
	ctx := context.Background()

	dto, err := svc.GetCourseDetail(ctx, nil, 1)
	if err != nil {
		t.Fatalf("GetCourseDetail: %v", err)
	}
	if dto.NotificationLevel != int(course.NotificationLevelNormal) {
		t.Errorf("NotificationLevel: got %d, want %d", dto.NotificationLevel, course.NotificationLevelNormal)
	}
}

func TestCourseQueryService_GetCourseDetail_WithTeacherGroup(t *testing.T) {
	query := newFakeCourseQuery()
	query.details[1] = &course.CourseDetailView{
		ID:           1,
		Code:         "CS101",
		Name:         "数据结构",
		LastSemester: "2025-2026-1",
		TeacherIDs:   []int{2, 1},
		MainTeacher:  &teacher.TeacherView{ID: 1, Name: "张三"},
	}
	teacherQuery := newFakeTeacherQuery()
	teacherQuery.views[1] = teacher.TeacherView{ID: 1, Name: "张三"}
	teacherQuery.views[2] = teacher.TeacherView{ID: 2, Name: "李四"}
	svc := application.NewCourseQueryService(query, teacherQuery, newFakeReviewQuery(), newFakeNotificationRepo(), nil)

	dto, err := svc.GetCourseDetail(context.Background(), nil, 1)
	if err != nil {
		t.Fatalf("GetCourseDetail: %v", err)
	}
	if dto.LastSemester != "2025-2026-1" {
		t.Errorf("LastSemester: got %q, want 2025-2026-1", dto.LastSemester)
	}
	if len(dto.TeacherGroup) != 2 {
		t.Fatalf("TeacherGroup count: got %d, want 2", len(dto.TeacherGroup))
	}
	if dto.TeacherGroup[0].ID != 2 || dto.TeacherGroup[1].ID != 1 {
		t.Errorf("TeacherGroup order: got [%d %d], want [2 1]", dto.TeacherGroup[0].ID, dto.TeacherGroup[1].ID)
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

	svc := application.NewCourseQueryService(query, nil, newFakeReviewQuery(), notifRepo, nil)

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
	svc := application.NewCourseQueryService(query, nil, newFakeReviewQuery(), newFakeNotificationRepo(), hotRepo)

	result, err := svc.ListHotCourses(context.Background(), "week", 5)
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

func TestCourseQueryService_ListHotCourses_WithLimit(t *testing.T) {
	query := newFakeCourseQuery()
	query.views[1] = course.CourseView{ID: 1, Code: "CS101", Name: "数据结构"}
	query.views[2] = course.CourseView{ID: 2, Code: "CS102", Name: "算法"}

	hotRepo := &fakeHotCourseRepo{ranks: []course.HotCourseRank{
		{CourseID: 2, Score: 10},
		{CourseID: 1, Score: 8},
	}}
	svc := application.NewCourseQueryService(query, nil, newFakeReviewQuery(), newFakeNotificationRepo(), hotRepo)

	t.Run("limit 1 returns only top course", func(t *testing.T) {
		result, err := svc.ListHotCourses(context.Background(), "week", 1)
		if err != nil {
			t.Fatalf("ListHotCourses: %v", err)
		}
		if len(result.Items) != 1 {
			t.Fatalf("items length: got %d, want 1", len(result.Items))
		}
		if result.Items[0].Course.ID != 2 {
			t.Fatalf("first item: got course=%d, want course=2", result.Items[0].Course.ID)
		}
	})
}

func TestCourseQueryService_ListHotCourses_InvalidPeriod(t *testing.T) {
	svc := application.NewCourseQueryService(newFakeCourseQuery(), nil, newFakeReviewQuery(), newFakeNotificationRepo(), &fakeHotCourseRepo{})
	_, err := svc.ListHotCourses(context.Background(), "daily", 5)
	if err != course.ErrInvalidHotCoursePeriod {
		t.Fatalf("err: got %v, want %v", err, course.ErrInvalidHotCoursePeriod)
	}
}
