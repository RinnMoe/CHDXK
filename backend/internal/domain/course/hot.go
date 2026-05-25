package course

import (
	"context"
	"encoding/json"
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

type HotCourseActivity string

const (
	HotCourseActivityReviewCreate HotCourseActivity = "review_create"
	HotCourseActivityReviewUpdate HotCourseActivity = "review_update"
	HotCourseActivityReviewVote   HotCourseActivity = "review_vote"
)

var ErrInvalidHotCourseActivity = errors.New("invalid hot course activity")

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
	AddScore(ctx context.Context, courseID int, score int64, at time.Time) error
	Top(ctx context.Context, period HotCoursePeriod, at time.Time, limit int64) ([]HotCourseRank, error)
}
