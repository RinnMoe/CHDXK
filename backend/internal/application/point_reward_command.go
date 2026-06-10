package application

import (
	"context"
	"strconv"
	"time"

	"jcourse/internal/domain/point"
)

type PointRewardCommandService struct {
	repo point.RewardRepository
}

func NewPointRewardCommandService(repo point.RewardRepository) *PointRewardCommandService {
	return &PointRewardCommandService{repo: repo}
}

func (s *PointRewardCommandService) GrantReward(ctx context.Context, rewardID int) error {
	return s.repo.GrantReward(ctx, rewardID, time.Now())
}

func (s *PointRewardCommandService) RevokeReviewRewardsByID(ctx context.Context, reviewID, courseID, authorUserID int) error {
	sources := []point.RewardSource{
		{
			UserID:     authorUserID,
			Reason:     point.RewardReasonReviewCreate,
			SourceType: point.RewardSourceTypeReview,
			SourceKey:  strconv.Itoa(reviewID),
		},
		{
			UserID:     authorUserID,
			Reason:     point.RewardReasonCourseFirstReview,
			SourceType: point.RewardSourceTypeCourse,
			SourceKey:  strconv.Itoa(courseID),
		},
	}
	return s.repo.RevokeRewardsBySources(ctx, sources, time.Now())
}
