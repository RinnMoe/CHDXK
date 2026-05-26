package course

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"jcourse/pkg/apperr"
)

type HotCoursePeriodName string

const (
	HotCoursePeriodWeek  HotCoursePeriodName = "week"
	HotCoursePeriodMonth HotCoursePeriodName = "month"
)

var ErrInvalidHotCoursePeriod = apperr.ErrInvalidHotCoursePeriod

type HotCoursePeriod struct {
	Period    HotCoursePeriodName
	PeriodKey string
}

func NewHotCoursePeriod(period HotCoursePeriodName, at time.Time, loc *time.Location) (HotCoursePeriod, error) {
	if period != HotCoursePeriodWeek && period != HotCoursePeriodMonth {
		return HotCoursePeriod{}, ErrInvalidHotCoursePeriod
	}
	return HotCoursePeriod{
		Period:    period,
		PeriodKey: HotCoursePeriodKey(period, at, loc),
	}, nil
}

func CurrentHotCoursePeriods(at time.Time, loc *time.Location) []HotCoursePeriod {
	return []HotCoursePeriod{
		{Period: HotCoursePeriodWeek, PeriodKey: HotCoursePeriodKey(HotCoursePeriodWeek, at, loc)},
		{Period: HotCoursePeriodMonth, PeriodKey: HotCoursePeriodKey(HotCoursePeriodMonth, at, loc)},
	}
}

type HotCourseRank struct {
	CourseID int
	Score    int64
}

const DefaultHotCourseLocationName = "Asia/Shanghai"

func DefaultHotCourseLocation() (*time.Location, error) {
	return time.LoadLocation(DefaultHotCourseLocationName)
}

func HotCoursePeriodKey(period HotCoursePeriodName, at time.Time, loc *time.Location) string {
	switch period {
	case HotCoursePeriodMonth:
		return hotCourseMonthKey(at, loc)
	default:
		return hotCourseWeekKey(at, loc)
	}
}

func HotCoursePeriodRange(period HotCoursePeriodName, at time.Time, loc *time.Location) (time.Time, time.Time) {
	switch period {
	case HotCoursePeriodMonth:
		start := hotCourseStartOfMonth(at, loc)
		return start, start.AddDate(0, 1, 0)
	default:
		start := hotCourseStartOfWeek(at, loc)
		return start, start.AddDate(0, 0, 7)
	}
}

func hotCourseMonthKey(at time.Time, loc *time.Location) string {
	t := at.In(hotCourseLocation(loc))
	return fmt.Sprintf("%04d-%02d", t.Year(), int(t.Month()))
}

func hotCourseWeekKey(at time.Time, loc *time.Location) string {
	year, week := at.In(hotCourseLocation(loc)).ISOWeek()
	return fmt.Sprintf("%04d-%02d", year, week)
}

func hotCourseStartOfMonth(at time.Time, loc *time.Location) time.Time {
	location := hotCourseLocation(loc)
	t := at.In(location)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, location)
}

func hotCourseStartOfWeek(at time.Time, loc *time.Location) time.Time {
	location := hotCourseLocation(loc)
	t := at.In(location)
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
	daysSinceMonday := (int(date.Weekday()) + 6) % 7
	return date.AddDate(0, 0, -daysSinceMonday)
}

func hotCourseLocation(loc *time.Location) *time.Location {
	if loc == nil {
		return time.UTC
	}
	return loc
}

type HotScoreConfig struct {
	ReviewCreateScore int64
	ReviewUpdateScore int64
	ReviewVoteScore   int64
}

var DefaultHotScoreConfig = HotScoreConfig{
	ReviewCreateScore: 3,
	ReviewUpdateScore: 1,
	ReviewVoteScore:   1,
}

type HotCourseActivity string

const (
	HotCourseActivityReviewCreate HotCourseActivity = "review_create"
	HotCourseActivityReviewUpdate HotCourseActivity = "review_update"
	HotCourseActivityReviewVote   HotCourseActivity = "review_vote"
)

var ErrInvalidHotCourseActivity = apperr.ErrInvalidHotCourseActivity

func (c HotScoreConfig) ScoreForActivity(activity HotCourseActivity) (int64, error) {
	switch activity {
	case HotCourseActivityReviewCreate:
		return c.ReviewCreateScore, nil
	case HotCourseActivityReviewUpdate:
		return c.ReviewUpdateScore, nil
	case HotCourseActivityReviewVote:
		return c.ReviewVoteScore, nil
	default:
		return 0, ErrInvalidHotCourseActivity
	}
}

const TaskTypeRecordHotCourseActivity = "course:record_hot_course_activity"

type RecordHotCourseActivityPayload struct {
	UserID   int               `json:"user_id"`
	Activity HotCourseActivity `json:"activity"`
	CourseID int               `json:"course_id"`
}

type RecordHotCourseActivityTask struct {
	payload RecordHotCourseActivityPayload
}

func NewRecordHotCourseActivityTask(userID int, activity HotCourseActivity, courseID int) RecordHotCourseActivityTask {
	return RecordHotCourseActivityTask{payload: RecordHotCourseActivityPayload{
		UserID:   userID,
		Activity: activity,
		CourseID: courseID,
	}}
}

func (t RecordHotCourseActivityTask) Type() string {
	return TaskTypeRecordHotCourseActivity
}

func (t RecordHotCourseActivityTask) Payload() []byte {
	b, _ := json.Marshal(t.payload)
	return b
}

type HotCourseRepository interface {
	AddScore(ctx context.Context, courseID int, score int64, periods ...HotCoursePeriod) error
	Top(ctx context.Context, period HotCoursePeriod, limit int64) ([]HotCourseRank, error)
}
