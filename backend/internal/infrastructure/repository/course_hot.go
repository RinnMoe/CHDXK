package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"jcourse/internal/domain/course"
)

const courseHotKeyPrefix = "course:hot"

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
		return courseHotKeyPrefix + ":month:" + monthKeyPart(at, r.loc)
	default:
		return courseHotKeyPrefix + ":week:" + weekKeyPart(at, r.loc)
	}
}

func monthKeyPart(at time.Time, loc *time.Location) string {
	t := at.In(loc)
	return fmt.Sprintf("%04d-%02d", t.Year(), int(t.Month()))
}

func weekKeyPart(at time.Time, loc *time.Location) string {
	monday := startOfWeek(at, loc)
	firstMonday := firstMondayOfYear(monday.Year(), loc)
	week := int(monday.Sub(firstMonday).Hours()/(24*7)) + 1
	return fmt.Sprintf("%04d-%02d", monday.Year(), week)
}

func startOfWeek(at time.Time, loc *time.Location) time.Time {
	t := at.In(loc)
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	daysSinceMonday := (int(date.Weekday()) + 6) % 7
	return date.AddDate(0, 0, -daysSinceMonday)
}

func firstMondayOfYear(year int, loc *time.Location) time.Time {
	date := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	daysUntilMonday := (int(time.Monday) - int(date.Weekday()) + 7) % 7
	return date.AddDate(0, 0, daysUntilMonday)
}

var _ course.HotCourseRepository = (*CourseHotRepository)(nil)
