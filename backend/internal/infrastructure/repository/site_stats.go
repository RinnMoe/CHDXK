package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/stat"
)

type SiteDailyStatRepository struct {
	db *gorm.DB
}

func NewSiteDailyStatRepository(db *gorm.DB) *SiteDailyStatRepository {
	return &SiteDailyStatRepository{db: db}
}

func metricsToJSONMap(metrics stat.Metrics) datatypes.JSONMap {
	if metrics == nil {
		return datatypes.JSONMap{}
	}
	jm := make(datatypes.JSONMap, len(metrics))
	for k, v := range metrics {
		jm[k] = v
	}
	return jm
}

func jsonMapToMetrics(metrics datatypes.JSONMap) stat.Metrics {
	if metrics == nil {
		return stat.Metrics{}
	}
	result := make(stat.Metrics, len(metrics))
	for k, v := range metrics {
		switch n := v.(type) {
		case int64:
			result[k] = n
		case int:
			result[k] = int64(n)
		case float64:
			result[k] = int64(n)
		case json.Number:
			if iv, err := n.Int64(); err == nil {
				result[k] = iv
			}
		}
	}
	return result
}

func newDailyStatView(e *SiteDailyStatEntity) stat.DailyStatView {
	metrics := jsonMapToMetrics(e.Metrics)
	return stat.DailyStatView{
		StatDate:            e.StatDate,
		TotalUserCount:      metrics[stat.MetricTotalUserCount],
		TotalReviewCount:    metrics[stat.MetricTotalReviewCount],
		ActiveUserCount:     metrics[stat.MetricActiveUserCount],
		NewUserCount:        metrics[stat.MetricNewUserCount],
		NewReviewCount:      metrics[stat.MetricNewReviewCount],
		ReviewAuthorCount:   metrics[stat.MetricReviewAuthorCount],
		ReviewedCourseTotal: metrics[stat.MetricReviewedCourseTotal],
		NewLikeCount:        metrics[stat.MetricNewLikeCount],
		NewDislikeCount:     metrics[stat.MetricNewDislikeCount],
		GeneratedAt:         e.GeneratedAt,
		UpdatedAt:           e.UpdatedAt,
	}
}

func newDailyStatEntity(s *stat.DailyStat) SiteDailyStatEntity {
	return SiteDailyStatEntity{
		StatDate:    s.StatDate,
		Metrics:     metricsToJSONMap(s.Metrics),
		GeneratedAt: s.GeneratedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func (r *SiteDailyStatRepository) Collect(ctx context.Context, periodStart, periodEnd time.Time) (stat.Metrics, error) {
	var totalUsers int64
	if err := r.db.WithContext(ctx).Model(&UserEntity{}).
		Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	var totalReviews int64
	if err := r.db.WithContext(ctx).Model(&ReviewEntity{}).
		Count(&totalReviews).Error; err != nil {
		return nil, err
	}

	var activeUsers int64
	if err := r.db.WithContext(ctx).Model(&UserEntity{}).
		Where("last_seen_at >= ? AND last_seen_at < ?", periodStart, periodEnd).
		Count(&activeUsers).Error; err != nil {
		return nil, err
	}

	var newUsers int64
	if err := r.db.WithContext(ctx).Model(&UserEntity{}).
		Where("created_at >= ? AND created_at < ?", periodStart, periodEnd).
		Count(&newUsers).Error; err != nil {
		return nil, err
	}

	var newReviews int64
	if err := r.db.WithContext(ctx).Model(&ReviewEntity{}).
		Where("created_at >= ? AND created_at < ?", periodStart, periodEnd).
		Count(&newReviews).Error; err != nil {
		return nil, err
	}

	var reviewAuthors int64
	if err := r.db.WithContext(ctx).Model(&ReviewEntity{}).
		Where("created_at >= ? AND created_at < ?", periodStart, periodEnd).
		Distinct("user_id").Count(&reviewAuthors).Error; err != nil {
		return nil, err
	}

	var reviewedCourses int64
	if err := r.db.WithContext(ctx).Model(&ReviewEntity{}).
		Where("created_at < ?", periodEnd).
		Distinct("course_id").Count(&reviewedCourses).Error; err != nil {
		return nil, err
	}

	var newLikes int64
	if err := r.db.WithContext(ctx).Model(&ReviewVoteEntity{}).
		Where("updated_at >= ? AND updated_at < ? AND vote_type = ?", periodStart, periodEnd, 1).
		Count(&newLikes).Error; err != nil {
		return nil, err
	}

	var newDislikes int64
	if err := r.db.WithContext(ctx).Model(&ReviewVoteEntity{}).
		Where("updated_at >= ? AND updated_at < ? AND vote_type = ?", periodStart, periodEnd, -1).
		Count(&newDislikes).Error; err != nil {
		return nil, err
	}

	return stat.Metrics{
		stat.MetricTotalUserCount:      totalUsers,
		stat.MetricTotalReviewCount:    totalReviews,
		stat.MetricActiveUserCount:     activeUsers,
		stat.MetricNewUserCount:        newUsers,
		stat.MetricNewReviewCount:      newReviews,
		stat.MetricReviewAuthorCount:   reviewAuthors,
		stat.MetricReviewedCourseTotal: reviewedCourses,
		stat.MetricNewLikeCount:        newLikes,
		stat.MetricNewDislikeCount:     newDislikes,
	}, nil
}

func (r *SiteDailyStatRepository) Upsert(ctx context.Context, s *stat.DailyStat) error {
	e := newDailyStatEntity(s)
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "stat_date"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"metrics",
				"generated_at",
				"updated_at",
			}),
		}).Create(&e).Error
}

func (r *SiteDailyStatRepository) GetByDate(ctx context.Context, statDate time.Time) (*stat.DailyStatView, error) {
	e, err := gorm.G[SiteDailyStatEntity](r.db).Where("stat_date = ?", statDate).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	v := newDailyStatView(&e)
	return &v, nil
}

func (r *SiteDailyStatRepository) FindByDateRange(ctx context.Context, filter stat.DailyStatFilter) ([]stat.DailyStatView, int64, error) {
	db := r.db.WithContext(ctx).Model(&SiteDailyStatEntity{}).
		Where("stat_date >= ? AND stat_date <= ?", filter.StartDate, filter.EndDate)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	db = db.Order(clause.OrderByColumn{
		Column: clause.Column{Table: "site_daily_stats", Name: "stat_date"},
		Desc:   true,
	})
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		db = db.Offset(offset).Limit(filter.PageSize)
	}

	var entities []SiteDailyStatEntity
	if err := db.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	items := make([]stat.DailyStatView, len(entities))
	for i, e := range entities {
		items[i] = newDailyStatView(&e)
	}
	return items, total, nil
}

var _ stat.DailyStatCollector = (*SiteDailyStatRepository)(nil)
var _ stat.DailyStatCommandRepository = (*SiteDailyStatRepository)(nil)
var _ stat.DailyStatQuery = (*SiteDailyStatRepository)(nil)
