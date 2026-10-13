package application_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"jcourse/internal/application"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type homeReviewQuery struct {
	review.ReviewQuery
	filter review.ReviewFilter
	calls  int
	err    error
}

func (q *homeReviewQuery) FindBy(ctx context.Context, filter review.ReviewFilter) ([]review.ReviewView, int64, error) {
	q.filter = filter
	q.calls++
	return nil, 0, q.err
}

type homeEnrollmentQuery struct {
	course.CourseEnrollmentQuery
	err error
}

func (q *homeEnrollmentQuery) FindUserEnrollments(context.Context, int) ([]course.CourseEnrollmentView, error) {
	return nil, q.err
}

func TestReviewPromptsUsesOneBatchAndPropagatesFailures(t *testing.T) {
	rows := &fakeEnrollmentQuery{rows: []course.CourseEnrollmentView{
		{UserID: 7, Course: course.CourseView{ID: 1}},
		{UserID: 7, Course: course.CourseView{ID: 2}},
		{UserID: 7, Course: course.CourseView{ID: 1}},
	}}
	reviews := &homeReviewQuery{}
	svc := application.NewCourseQueryService(nil, nil, reviews, nil, rows, nil)
	if _, err := svc.ReviewPrompts(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if reviews.calls != 1 || reviews.filter.UserID != 7 || len(reviews.filter.CourseIDs) != 2 || reviews.filter.PageSize != 0 {
		t.Fatalf("expected single unpaginated batch, got %+v (%d calls)", reviews.filter, reviews.calls)
	}
	failure := errors.New("storage unavailable")
	reviews.err = failure
	if _, err := svc.ReviewPrompts(context.Background(), 7); !errors.Is(err, failure) {
		t.Fatal("review failure was hidden", err)
	}
	svc = application.NewCourseQueryService(nil, nil, reviews, nil, &homeEnrollmentQuery{err: failure}, nil)
	if _, err := svc.ReviewPrompts(context.Background(), 7); !errors.Is(err, failure) {
		t.Fatal("enrollment failure was hidden", err)
	}
}

func TestDiscoverCoursesUsesLatestRecognizedSemester(t *testing.T) {
	for _, tt := range []struct {
		name      string
		semesters []string
		want      string
	}{
		{"unsorted", []string{"2025-2026-2", "unknown", "2026-2027-1", "2026-2027-2", "2027-2026-1"}, "2026-2027-2"},
		{"unrecognized", []string{"unknown", "2026-2028-1", "2026-2027-3"}, ""},
		{"empty", nil, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := course.NewMockCourseRepository()
			repo.Filters = &course.CourseFilters{}
			for _, name := range tt.semesters {
				repo.Filters.Semesters = append(repo.Filters.Semesters, course.FilterItem{Name: name})
			}
			repo.OnFindDiscovery = func(_ context.Context, semester, date string, limit int) ([]course.CourseView, error) {
				loc, _ := time.LoadLocation("Asia/Shanghai")
				if semester != tt.want || date != time.Now().In(loc).Format("2006-01-02") || limit != 24 {
					t.Fatalf("query = %q %q %d", semester, date, limit)
				}
				return []course.CourseView{{ID: 2}, {ID: 1}}, nil
			}
			svc := application.NewCourseQueryService(repo, nil, nil, nil, nil, nil)
			got, err := svc.DiscoverCourses(context.Background())
			if err != nil || got.Semester != tt.want || len(got.Items) != 2 || got.Items[0].ID != 2 {
				t.Fatalf("discovery = %+v, %v", got, err)
			}
		})
	}
}

func TestDiscoverCoursesPreservesEmptyAndErrors(t *testing.T) {
	repo := course.NewMockCourseRepository()
	svc := application.NewCourseQueryService(repo, nil, nil, nil, nil, nil)
	got, err := svc.DiscoverCourses(context.Background())
	if err != nil || got.Items == nil || len(got.Items) != 0 {
		t.Fatalf("empty = %+v, %v", got, err)
	}
	failure := errors.New("database unavailable")
	repo.OnGetFilters = func(context.Context) (*course.CourseFilters, error) { return nil, failure }
	if _, err := svc.DiscoverCourses(context.Background()); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	repo.OnGetFilters = nil
	repo.OnFindDiscovery = func(context.Context, string, string, int) ([]course.CourseView, error) { return nil, failure }
	if _, err := svc.DiscoverCourses(context.Background()); !errors.Is(err, failure) {
		t.Fatal(err)
	}
}

func TestReviewPromptsOrderingDeduplicationAndOwnership(t *testing.T) {
	rows := []course.CourseEnrollmentView{
		{UserID: 1, Course: course.CourseView{ID: 3}, Semester: "2025-2026-1"},
		{UserID: 1, Course: course.CourseView{ID: 2}, Semester: "2026-2027-1"},
		{UserID: 1, Course: course.CourseView{ID: 3}, Semester: "2026-2027-1"},
		{UserID: 1, Course: course.CourseView{ID: 1}, Semester: "2026-2027-1"},
		{UserID: 1, Course: course.CourseView{ID: 4}, Semester: "2025-2026-2"},
		{UserID: 1, Course: course.CourseView{ID: 5}, Semester: "2024-2025-1"},
		{UserID: 2, Course: course.CourseView{ID: 6}, Semester: "2026-2027-2"},
	}
	reviews := newFakeReviewQuery()
	reviews.reviews = []review.ReviewView{{UserID: 1, CourseID: 2}, {UserID: 2, CourseID: 1}}
	svc := application.NewCourseQueryService(nil, nil, reviews, nil, &fakeEnrollmentQuery{rows: rows}, nil)
	got, err := svc.ReviewPrompts(context.Background(), 1)
	if err != nil || !got.HasEnrollments {
		t.Fatalf("result = %+v, %v", got, err)
	}
	ids := []int{}
	for _, item := range got.Items {
		ids = append(ids, item.Course.ID)
	}
	if !reflect.DeepEqual(ids, []int{1, 3, 4}) || got.Items[1].Semester != "2026-2027-1" {
		t.Fatalf("unexpected prompts: %+v", got.Items)
	}
	// Source data is not sorted or mutated as a side effect.
	if rows[0].Semester != "2025-2026-1" {
		t.Fatal("mutated enrollment rows")
	}
	other, err := svc.ReviewPrompts(context.Background(), 2)
	if err != nil || len(other.Items) != 1 || other.Items[0].Course.ID != 6 {
		t.Fatalf("other user = %+v, %v", other, err)
	}
}

func TestReviewPromptsEmptyStates(t *testing.T) {
	for _, tt := range []struct {
		name     string
		enrolled bool
		reviewed bool
	}{
		{"no records", false, false}, {"all reviewed", true, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			enrollments := &fakeEnrollmentQuery{}
			reviews := newFakeReviewQuery()
			if tt.enrolled {
				enrollments.rows = []course.CourseEnrollmentView{{UserID: 1, Course: course.CourseView{ID: 1}}}
			}
			if tt.reviewed {
				reviews.reviews = []review.ReviewView{{UserID: 1, CourseID: 1}}
			}
			svc := application.NewCourseQueryService(nil, nil, reviews, nil, enrollments, nil)
			got, err := svc.ReviewPrompts(context.Background(), 1)
			if err != nil || got.HasEnrollments != tt.enrolled || got.Items == nil || len(got.Items) != 0 {
				t.Fatalf("result = %+v, %v", got, err)
			}
		})
	}
}
