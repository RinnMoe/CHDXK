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
)

const DefaultHotCourseLocationName = "Asia/Shanghai"

func DefaultHotCourseLocation() (*time.Location, error) {
	return time.LoadLocation(DefaultHotCourseLocationName)
}

type CourseHotRepository struct {
	client *redis.Client
	loc    *time.Location
}

func NewCourseHotRepository(client *redis.Client, loc *time.Location) *CourseHotRepository {
	if loc == nil {
		loc = time.UTC
	}
	return &CourseHotRepository{client: client, loc: loc}
}

func (r *CourseHotRepository) AddScore(ctx context.Context, courseID int, score int64, at time.Time) error {
	if score == 0 {
		return nil
	}

	member := strconv.Itoa(courseID)
	pipe := r.client.Pipeline()
	pipe.ZIncrBy(ctx, r.key(course.HotCoursePeriodWeek, at), float64(score), member)
	pipe.ZIncrBy(ctx, r.key(course.HotCoursePeriodMonth, at), float64(score), member)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *CourseHotRepository) Top(ctx context.Context, period course.HotCoursePeriod, at time.Time, limit int64) ([]course.HotCourseRank, error) {
	if limit <= 0 {
		return []course.HotCourseRank{}, nil
	}
	if period != course.HotCoursePeriodWeek && period != course.HotCoursePeriodMonth {
		return nil, course.ErrInvalidHotCoursePeriod
	}

	items, err := r.client.ZRevRangeWithScores(ctx, r.key(period, at), 0, limit-1).Result()
	if err != nil {
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

func (r *CourseHotRepository) key(period course.HotCoursePeriod, at time.Time) string {
	switch period {
	case course.HotCoursePeriodMonth:
		return redisKey("course", "hot", "month", HotCoursePeriodKey(course.HotCoursePeriodMonth, at, r.loc))
	default:
		return redisKey("course", "hot", "week", HotCoursePeriodKey(course.HotCoursePeriodWeek, at, r.loc))
	}
}

type GormCourseHotRepository struct {
	db    *gorm.DB
	loc   *time.Location
	cache *redis.Client
}

func NewGormCourseHotRepository(db *gorm.DB, cache ...*redis.Client) *GormCourseHotRepository {
	loc, err := DefaultHotCourseLocation()
	if err != nil {
		panic(err)
	}
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &GormCourseHotRepository{db: db, loc: loc, cache: client}
}

func (r *GormCourseHotRepository) AddScore(ctx context.Context, courseID int, score int64, at time.Time) error {
	if score == 0 {
		return nil
	}

	now := time.Now()
	items := []CourseHotScoreEntity{
		r.newScoreEntity(course.HotCoursePeriodWeek, courseID, score, at, now),
		r.newScoreEntity(course.HotCoursePeriodMonth, courseID, score, at, now),
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "period"},
			{Name: "period_key"},
			{Name: "course_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"score":      gorm.Expr("course_hot_scores.score + EXCLUDED.score"),
			"updated_at": now,
		}),
	}).Create(&items).Error; err != nil {
		return err
	}
	cacheDeletePattern(ctx, r.cache, cacheKey("course", "hot", "*")+":*")
	return nil
}

func (r *GormCourseHotRepository) Top(ctx context.Context, period course.HotCoursePeriod, at time.Time, limit int64) ([]course.HotCourseRank, error) {
	if limit <= 0 {
		return []course.HotCourseRank{}, nil
	}
	if period != course.HotCoursePeriodWeek && period != course.HotCoursePeriodMonth {
		return nil, course.ErrInvalidHotCoursePeriod
	}

	periodKey := HotCoursePeriodKey(period, at, r.loc)
	key := cacheKey("course", "hot", period, periodKey, limit)
	if cached, ok := cacheGetJSON[[]course.HotCourseRank](ctx, r.cache, key); ok {
		return *cached, nil
	}

	var items []CourseHotScoreEntity
	if err := r.db.WithContext(ctx).
		Where("period = ? AND period_key = ?", string(period), periodKey).
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

func (r *GormCourseHotRepository) newScoreEntity(period course.HotCoursePeriod, courseID int, score int64, at time.Time, now time.Time) CourseHotScoreEntity {
	return CourseHotScoreEntity{
		Period:    string(period),
		PeriodKey: HotCoursePeriodKey(period, at, r.loc),
		CourseID:  courseID,
		Score:     score,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func HotCoursePeriodKey(period course.HotCoursePeriod, at time.Time, loc *time.Location) string {
	switch period {
	case course.HotCoursePeriodMonth:
		return monthKeyPart(at, loc)
	default:
		return weekKeyPart(at, loc)
	}
}

func HotCoursePeriodRange(period course.HotCoursePeriod, at time.Time, loc *time.Location) (time.Time, time.Time) {
	switch period {
	case course.HotCoursePeriodMonth:
		start := startOfHotCourseMonth(at, loc)
		return start, start.AddDate(0, 1, 0)
	default:
		start := hotCourseStartOfWeek(at, loc)
		return start, start.AddDate(0, 0, 7)
	}
}

func monthKeyPart(at time.Time, loc *time.Location) string {
	t := at.In(loc)
	return fmt.Sprintf("%04d-%02d", t.Year(), int(t.Month()))
}

func startOfHotCourseMonth(at time.Time, loc *time.Location) time.Time {
	t := at.In(loc)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
}

func weekKeyPart(at time.Time, loc *time.Location) string {
	monday := hotCourseStartOfWeek(at, loc)
	firstMonday := hotCourseFirstMondayOfYear(monday.Year(), loc)
	week := int(monday.Sub(firstMonday).Hours()/(24*7)) + 1
	return fmt.Sprintf("%04d-%02d", monday.Year(), week)
}

func hotCourseStartOfWeek(at time.Time, loc *time.Location) time.Time {
	t := at.In(loc)
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	daysSinceMonday := (int(date.Weekday()) + 6) % 7
	return date.AddDate(0, 0, -daysSinceMonday)
}

func hotCourseFirstMondayOfYear(year int, loc *time.Location) time.Time {
	date := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	daysUntilMonday := (int(time.Monday) - int(date.Weekday()) + 7) % 7
	return date.AddDate(0, 0, daysUntilMonday)
}

var _ course.HotCourseRepository = (*CourseHotRepository)(nil)
var _ course.HotCourseRepository = (*GormCourseHotRepository)(nil)
