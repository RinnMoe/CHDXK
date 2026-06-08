package application

import (
	"time"

	"jcourse/internal/domain/stat"
)

type SiteDailyStatDTO struct {
	StatDate            string    `json:"stat_date"`
	ActiveUserCount     int64     `json:"active_user_count"`
	NewUserCount        int64     `json:"new_user_count"`
	NewReviewCount      int64     `json:"new_review_count"`
	NewPointAmount      int64     `json:"new_point_amount"`
	ReviewAuthorCount   int64     `json:"review_author_count"`
	NewLikeCount        int64     `json:"new_like_count"`
	NewDislikeCount     int64     `json:"new_dislike_count"`
	TotalUserCount      int64     `json:"total_user_count"`
	TotalReviewCount    int64     `json:"total_review_count"`
	ReviewedCourseTotal int64     `json:"reviewed_course_total"`
	GeneratedAt         time.Time `json:"generated_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func newSiteDailyStatDTO(s *stat.DailyStat) SiteDailyStatDTO {
	view := stat.DailyStatView{
		StatDate:            s.StatDate,
		ActiveUserCount:     s.Metrics[stat.MetricActiveUserCount],
		NewUserCount:        s.Metrics[stat.MetricNewUserCount],
		NewReviewCount:      s.Metrics[stat.MetricNewReviewCount],
		NewPointAmount:      s.Metrics[stat.MetricNewPointAmount],
		ReviewAuthorCount:   s.Metrics[stat.MetricReviewAuthorCount],
		NewLikeCount:        s.Metrics[stat.MetricNewLikeCount],
		NewDislikeCount:     s.Metrics[stat.MetricNewDislikeCount],
		TotalUserCount:      s.Metrics[stat.MetricTotalUserCount],
		TotalReviewCount:    s.Metrics[stat.MetricTotalReviewCount],
		ReviewedCourseTotal: s.Metrics[stat.MetricReviewedCourseTotal],
		GeneratedAt:         s.GeneratedAt,
		UpdatedAt:           s.UpdatedAt,
	}
	return newSiteDailyStatViewDTO(&view)
}

func newSiteDailyStatViewDTO(s *stat.DailyStatView) SiteDailyStatDTO {
	return SiteDailyStatDTO{
		StatDate:            s.StatDate.Format(stat.DateLayout),
		ActiveUserCount:     s.ActiveUserCount,
		NewUserCount:        s.NewUserCount,
		NewReviewCount:      s.NewReviewCount,
		NewPointAmount:      s.NewPointAmount,
		ReviewAuthorCount:   s.ReviewAuthorCount,
		NewLikeCount:        s.NewLikeCount,
		NewDislikeCount:     s.NewDislikeCount,
		TotalUserCount:      s.TotalUserCount,
		TotalReviewCount:    s.TotalReviewCount,
		ReviewedCourseTotal: s.ReviewedCourseTotal,
		GeneratedAt:         s.GeneratedAt,
		UpdatedAt:           s.UpdatedAt,
	}
}
