package stat

import (
	"context"
	"time"
)

const (
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

type DailyStatView struct {
	StatDate            time.Time
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
