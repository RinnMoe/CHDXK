package stat

import (
	"context"
	"errors"
	"time"
)

const DateLayout = "2006-01-02"

const DefaultTimezoneName = "Asia/Shanghai"

type Config struct {
	DailyCron        string
	SchedulerEnabled bool
	Timezone         string
}

var DefaultConfig = Config{
	DailyCron:        "10 0 * * *",
	SchedulerEnabled: true,
	Timezone:         DefaultTimezoneName,
}

var ErrInvalidDateRange = errors.New("invalid date range")

const (
	MetricTotalUserCount      = "total_user_count"
	MetricTotalReviewCount    = "total_review_count"
	MetricActiveUserCount     = "active_user_count"
	MetricNewUserCount        = "new_user_count"
	MetricNewReviewCount      = "new_review_count"
	MetricReviewAuthorCount   = "review_author_count"
	MetricReviewedCourseTotal = "reviewed_course_total"
	MetricNewLikeCount        = "new_like_count"
	MetricNewDislikeCount     = "new_dislike_count"
)

type Metrics map[string]int64

type DailyStat struct {
	StatDate    time.Time
	Metrics     Metrics
	GeneratedAt time.Time
	UpdatedAt   time.Time
}

type Calendar struct {
	loc *time.Location
}

func NewCalendar(loc *time.Location) Calendar {
	if loc == nil {
		loc = time.UTC
	}
	return Calendar{loc: loc}
}

func (c Calendar) DateOnly(value time.Time) time.Time {
	v := value.In(c.loc)
	return time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, c.loc)
}

func (c Calendar) Yesterday() time.Time {
	return c.DateOnly(time.Now().In(c.loc).AddDate(0, 0, -1))
}

func (c Calendar) ParseDate(value string) (time.Time, error) {
	date, err := time.ParseInLocation(DateLayout, value, c.loc)
	if err != nil {
		return time.Time{}, err
	}
	return c.DateOnly(date), nil
}

func (c Calendar) ParseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	if startDate == "" || endDate == "" {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	start, err := c.ParseDate(startDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := c.ParseDate(endDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	return start, end, nil
}

type DailyStatService struct {
	collector DailyStatCollector
	repo      DailyStatCommandRepository
	calendar  Calendar
}

func NewDailyStatService(collector DailyStatCollector, repo DailyStatCommandRepository, loc *time.Location) *DailyStatService {
	return &DailyStatService{collector: collector, repo: repo, calendar: NewCalendar(loc)}
}

func (s *DailyStatService) CollectDailyByDateString(ctx context.Context, statDate string) (*DailyStat, error) {
	if statDate == "" {
		return s.CollectDaily(ctx, s.calendar.Yesterday())
	}
	date, err := s.calendar.ParseDate(statDate)
	if err != nil {
		return nil, err
	}
	return s.CollectDaily(ctx, date)
}

func (s *DailyStatService) CollectDaily(ctx context.Context, statDate time.Time) (*DailyStat, error) {
	date := s.calendar.DateOnly(statDate)
	periodStart := date
	periodEnd := periodStart.AddDate(0, 0, 1)

	metrics, err := s.collector.Collect(ctx, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	now := time.Now().In(s.calendar.loc)
	daily := &DailyStat{
		StatDate:    date,
		Metrics:     metrics,
		GeneratedAt: now,
		UpdatedAt:   now,
	}
	if err := s.repo.Upsert(ctx, daily); err != nil {
		return nil, err
	}
	return daily, nil
}

type DailyStatView struct {
	StatDate            time.Time
	TotalUserCount      int64
	TotalReviewCount    int64
	ActiveUserCount     int64
	NewUserCount        int64
	NewReviewCount      int64
	ReviewAuthorCount   int64
	ReviewedCourseTotal int64
	NewLikeCount        int64
	NewDislikeCount     int64
	GeneratedAt         time.Time
	UpdatedAt           time.Time
}

type DailyStatFilter struct {
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PageSize  int
}

type DailyStatCollector interface {
	Collect(ctx context.Context, periodStart, periodEnd time.Time) (Metrics, error)
}

type DailyStatCommandRepository interface {
	Upsert(ctx context.Context, stat *DailyStat) error
}

type DailyStatQuery interface {
	GetByDate(ctx context.Context, statDate time.Time) (*DailyStatView, error)
	FindByDateRange(ctx context.Context, filter DailyStatFilter) ([]DailyStatView, int64, error)
}
