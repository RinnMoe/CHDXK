package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/course"
	"jcourse/pkg/logx"
)

type CourseHotRepository struct {
	client *redis.Client
}

func NewCourseHotRepository(client *redis.Client) *CourseHotRepository {
	return &CourseHotRepository{client: client}
}

func (r *CourseHotRepository) AddScore(ctx context.Context, courseID int, score int64, periods ...course.HotCoursePeriod) error {
	if score == 0 || len(periods) == 0 {
		return nil
	}

	member := strconv.Itoa(courseID)
	pipe := r.client.Pipeline()
	keys := make([]string, 0, len(periods))
	for _, period := range periods {
		key := r.key(period)
		keys = append(keys, key)
		pipe.ZIncrBy(ctx, key, float64(score), member)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		logx.Warn(ctx, "cache access failed", "operation", "zincrby_pipeline", "keys", keys, "err", err)
	}
	return err
}

func (r *CourseHotRepository) Top(ctx context.Context, period course.HotCoursePeriod, limit int64) ([]course.HotCourseRank, error) {
	if limit <= 0 {
		return []course.HotCourseRank{}, nil
	}
	if period.Period != course.HotCoursePeriodWeek && period.Period != course.HotCoursePeriodMonth {
		return nil, course.ErrInvalidHotCoursePeriod
	}

	key := r.key(period)
	items, err := r.client.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		logCacheAccessFailure(ctx, "zrevrange_with_scores", key, err)
		return nil, err
	}

	ranks := make([]course.HotCourseRank, 0, len(items))
	for _, item := range items {
		courseID, err := strconv.Atoi(fmt.Sprint(item.Member))
		if err != nil {
			continue
		}
		ranks = append(ranks, course.HotCourseRank{
			CourseID: courseID,
			Score:    int64(item.Score),
		})
	}
	return ranks, nil
}

func (r *CourseHotRepository) key(period course.HotCoursePeriod) string {
	return redisKey("course", "hot", period.Period, period.PeriodKey)
}

type GormCourseHotRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewGormCourseHotRepository(db *gorm.DB, cache ...*redis.Client) *GormCourseHotRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &GormCourseHotRepository{db: db, cache: client}
}

func (r *GormCourseHotRepository) AddScore(ctx context.Context, courseID int, score int64, periods ...course.HotCoursePeriod) error {
	if score == 0 || len(periods) == 0 {
		return nil
	}

	now := time.Now()
	items := make([]CourseHotScoreEntity, 0, len(periods))
	for _, period := range periods {
		items = append(items, r.newScoreEntity(period, courseID, score, now))
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "period"},
			{Name: "period_key"},
			{Name: "course_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"score":      gorm.Expr("course_hot_scores.score + EXCLUDED.score"),
			"updated_at": now,
		}),
	}).Create(&items).Error; err != nil {
		return err
	}
	for _, period := range periods {
		cacheDeletePattern(ctx, r.cache, cacheKey("course", "hot", period.Period, period.PeriodKey)+":*")
	}
	return nil
}

func (r *GormCourseHotRepository) Top(ctx context.Context, period course.HotCoursePeriod, limit int64) ([]course.HotCourseRank, error) {
	if limit <= 0 {
		return []course.HotCourseRank{}, nil
	}
	if period.Period != course.HotCoursePeriodWeek && period.Period != course.HotCoursePeriodMonth {
		return nil, course.ErrInvalidHotCoursePeriod
	}

	key := cacheKey("course", "hot", period.Period, period.PeriodKey, limit)
	if cached, ok := cacheGetJSON[[]course.HotCourseRank](ctx, r.cache, key); ok {
		return *cached, nil
	}

	var items []CourseHotScoreEntity
	if err := r.db.WithContext(ctx).
		Where("period = ? AND period_key = ?", string(period.Period), period.PeriodKey).
		Order("score DESC, course_id ASC").
		Limit(int(limit)).
		Find(&items).Error; err != nil {
		return nil, err
	}

	ranks := make([]course.HotCourseRank, 0, len(items))
	for _, item := range items {
		ranks = append(ranks, course.HotCourseRank{
			CourseID: item.CourseID,
			Score:    item.Score,
		})
	}
	cacheSetJSON(ctx, r.cache, key, ranks)
	return ranks, nil
}

func (r *GormCourseHotRepository) newScoreEntity(period course.HotCoursePeriod, courseID int, score int64, now time.Time) CourseHotScoreEntity {
	return CourseHotScoreEntity{
		Period:    string(period.Period),
		PeriodKey: period.PeriodKey,
		CourseID:  courseID,
		Score:     score,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

var _ course.HotCourseRepository = (*CourseHotRepository)(nil)
var _ course.HotCourseRepository = (*GormCourseHotRepository)(nil)
