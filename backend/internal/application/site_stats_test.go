package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/application"
	"jcourse/internal/domain/stat"
)

func TestSiteStatsCommandService_CollectDailyByDateString(t *testing.T) {
	collector := &fakeDailyStatCollector{metrics: stat.Metrics{
		stat.MetricActiveUserCount:     11,
		stat.MetricNewUserCount:        2,
		stat.MetricNewReviewCount:      5,
		stat.MetricReviewAuthorCount:   3,
		stat.MetricReviewedCourseTotal: 9,
		stat.MetricNewLikeCount:        7,
		stat.MetricNewDislikeCount:     1,
	}}
	repo := &stat.MockDailyStatCommandRepository{}
	svc := application.NewSiteStatsCommandService(collector, repo, stat.DefaultConfig)

	got, err := svc.CollectDailyByDateString(context.Background(), "2026-05-20")
	if err != nil {
		t.Fatalf("CollectDailyByDateString: %v", err)
	}

	loc := mustTestStatsLocation(t)
	wantStart := time.Date(2026, 5, 20, 0, 0, 0, 0, loc)
	wantEnd := wantStart.AddDate(0, 0, 1)
	if !collector.periodStart.Equal(wantStart) || !collector.periodEnd.Equal(wantEnd) {
		t.Fatalf("period = [%s, %s), want [%s, %s)", collector.periodStart, collector.periodEnd, wantStart, wantEnd)
	}
	if repo.Saved == nil {
		t.Fatal("expected stat to be saved")
	}
	if repo.Saved.StatDate.Format("2006-01-02") != "2026-05-20" {
		t.Fatalf("saved stat date = %s", repo.Saved.StatDate.Format("2006-01-02"))
	}
	if got.ActiveUserCount != 11 || got.ReviewAuthorCount != 3 || got.NewDislikeCount != 1 {
		t.Fatalf("flattened metrics = %+v", got)
	}
}

func TestSiteStatsQueryService_ListDaily(t *testing.T) {
	loc := mustTestStatsLocation(t)
	query := &fakeDailyStatQuery{
		items: []stat.DailyStatView{
			{StatDate: time.Date(2026, 5, 20, 0, 0, 0, 0, loc), ActiveUserCount: 5},
			{StatDate: time.Date(2026, 5, 19, 0, 0, 0, 0, loc), ActiveUserCount: 4},
		},
	}
	svc := application.NewSiteStatsQueryService(query, stat.DefaultConfig)

	got, err := svc.ListDaily(context.Background(), application.SiteDailyStatListFilter{
		StartDate: "2026-05-01",
		EndDate:   "2026-05-20",
	})
	if err != nil {
		t.Fatalf("ListDaily: %v", err)
	}

	if query.filter.StartDate.Format("2006-01-02") != "2026-05-01" || query.filter.EndDate.Format("2006-01-02") != "2026-05-20" {
		t.Fatalf("filter date range = %s..%s", query.filter.StartDate, query.filter.EndDate)
	}
	if len(got) != 2 || got[0].StatDate != "2026-05-20" || got[0].ActiveUserCount != 5 {
		t.Fatalf("result = %+v", got)
	}
}

func TestSiteStatsQueryService_ListDaily_InvalidRange(t *testing.T) {
	svc := application.NewSiteStatsQueryService(&fakeDailyStatQuery{}, stat.DefaultConfig)
	_, err := svc.ListDaily(context.Background(), application.SiteDailyStatListFilter{
		StartDate: "2026-05-20",
		EndDate:   "2026-05-01",
	})
	if !errors.Is(err, application.ErrInvalidDateRange) {
		t.Fatalf("error = %v, want ErrInvalidDateRange", err)
	}
}

type fakeDailyStatCollector struct {
	metrics     stat.Metrics
	periodStart time.Time
	periodEnd   time.Time
}

func (c *fakeDailyStatCollector) Collect(_ context.Context, periodStart, periodEnd time.Time) (stat.Metrics, error) {
	c.periodStart = periodStart
	c.periodEnd = periodEnd
	return c.metrics, nil
}

type fakeDailyStatQuery struct {
	filter stat.DailyStatFilter
	items  []stat.DailyStatView
}

func (q *fakeDailyStatQuery) GetByDate(_ context.Context, statDate time.Time) (*stat.DailyStatView, error) {
	return &stat.DailyStatView{StatDate: statDate}, nil
}

func (q *fakeDailyStatQuery) FindByDateRange(_ context.Context, filter stat.DailyStatFilter) ([]stat.DailyStatView, error) {
	q.filter = filter
	return q.items, nil
}

func mustTestStatsLocation(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	return loc
}
