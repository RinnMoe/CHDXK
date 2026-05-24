package course

import (
	"context"
	"errors"
	"time"
)

type HotCoursePeriod string

const (
	HotCoursePeriodWeek  HotCoursePeriod = "week"
	HotCoursePeriodMonth HotCoursePeriod = "month"
)

var ErrInvalidHotCoursePeriod = errors.New("invalid hot course period")

type HotCourseRank struct {
	CourseID int
	Score    int64
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

type HotCourseRepository interface {
	AddScore(ctx context.Context, courseID int, score int64, at time.Time) error
	Top(ctx context.Context, period HotCoursePeriod, at time.Time, limit int64) ([]HotCourseRank, error)
}
