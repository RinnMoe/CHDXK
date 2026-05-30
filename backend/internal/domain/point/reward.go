package point

import (
	"context"
	"strconv"
	"time"
)

type RewardReason string
type RewardStatus string

const (
	RecordReasonReward RecordReason = "reward"

	RewardReasonCourseFirstReview RewardReason = "course_first_review"

	RewardStatusPending  RewardStatus = "pending"
	RewardStatusGranted  RewardStatus = "granted"
	RewardStatusCanceled RewardStatus = "canceled"

	RewardSourceTypeCourse = "course"
)

type RewardConfig struct {
	Enabled                 bool `mapstructure:"enabled"`
	CourseFirstReviewPoints int  `mapstructure:"course_first_review_points"`
}

var DefaultRewardConfig = RewardConfig{
	Enabled:                 false,
	CourseFirstReviewPoints: 0,
}

type Reward struct {
	ID          int
	UserID      int
	Reason      RewardReason
	Amount      int
	SourceType  string
	SourceKey   string
	Description string
	Status      RewardStatus
	CreatedAt   time.Time
	GrantedAt   *time.Time
}

func NewCourseFirstReviewReward(userID int, courseID int, amount int, now time.Time) *Reward {
	if amount <= 0 {
		return nil
	}
	return &Reward{
		UserID:      userID,
		Reason:      RewardReasonCourseFirstReview,
		Amount:      amount,
		SourceType:  RewardSourceTypeCourse,
		SourceKey:   strconv.Itoa(courseID),
		Description: "课程首评奖励",
		Status:      RewardStatusPending,
		CreatedAt:   now,
	}
}

func (r *Reward) GrantRecord(now time.Time) Record {
	return Record{
		UserID:      r.UserID,
		Reason:      RecordReasonReward,
		Amount:      r.Amount,
		Description: r.Description,
		CreatedAt:   now,
	}
}

type RewardRepository interface {
	GrantReward(ctx context.Context, rewardID int, now time.Time) error
}
