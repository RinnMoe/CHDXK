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
	RewardReasonReviewCreate      RewardReason = "review_create"

	RewardStatusPending  RewardStatus = "pending"
	RewardStatusGranted  RewardStatus = "granted"
	RewardStatusCanceled RewardStatus = "canceled"

	RewardSourceTypeCourse = "course"
	RewardSourceTypeReview = "review"
)

type RewardConfig struct {
	CourseFirstReviewEnabled bool `mapstructure:"course_first_review_enabled"`
	CourseFirstReviewPoints  int  `mapstructure:"course_first_review_points"`
	ReviewCreateEnabled      bool `mapstructure:"review_create_enabled"`
	ReviewCreatePoints       int  `mapstructure:"review_create_points"`
}

var DefaultRewardConfig = RewardConfig{
	CourseFirstReviewEnabled: false,
	CourseFirstReviewPoints:  0,
	ReviewCreateEnabled:      false,
	ReviewCreatePoints:       1,
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

func NewReviewCreateReward(userID int, amount int, now time.Time) *Reward {
	if amount <= 0 {
		return nil
	}
	return &Reward{
		UserID:      userID,
		Reason:      RewardReasonReviewCreate,
		Amount:      amount,
		SourceType:  RewardSourceTypeReview,
		Description: "发布点评奖励",
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
