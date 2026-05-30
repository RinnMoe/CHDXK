package application

import (
	"context"
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
