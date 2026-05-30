package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/point"
)

func newPointRewardEntity(r *point.Reward) PointRewardEntity {
	return PointRewardEntity{
		ID:          r.ID,
		UserID:      r.UserID,
		Reason:      string(r.Reason),
		Amount:      r.Amount,
		SourceType:  r.SourceType,
		SourceKey:   r.SourceKey,
		Description: r.Description,
		Status:      string(r.Status),
		CreatedAt:   r.CreatedAt,
		GrantedAt:   r.GrantedAt,
	}
}

func newPointRewardDomain(e *PointRewardEntity) point.Reward {
	return point.Reward{
		ID:          e.ID,
		UserID:      e.UserID,
		Reason:      point.RewardReason(e.Reason),
		Amount:      e.Amount,
		SourceType:  e.SourceType,
		SourceKey:   e.SourceKey,
		Description: e.Description,
		Status:      point.RewardStatus(e.Status),
		CreatedAt:   e.CreatedAt,
		GrantedAt:   e.GrantedAt,
	}
}

func (r *PointRepository) GrantReward(ctx context.Context, rewardID int, now time.Time) error {
	var userID int
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var entity PointRewardEntity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", rewardID).Take(&entity).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		userID = entity.UserID

		reward := newPointRewardDomain(&entity)
		if reward.Status == point.RewardStatusGranted || reward.Status == point.RewardStatusCanceled {
			return nil
		}

		recordEntity := newPointRecordEntity(reward.GrantRecord(now))
		if err := tx.Create(&recordEntity).Error; err != nil {
			return err
		}

		return tx.Model(&PointRewardEntity{}).Where("id = ?", rewardID).Updates(map[string]any{
			"status":     string(point.RewardStatusGranted),
			"granted_at": now,
		}).Error
	}); err != nil {
		return err
	}
	if userID != 0 {
		cacheDelete(ctx, r.cache, cacheKey("point", userID, "sum"))
	}
	return nil
}

var _ point.RewardRepository = (*PointRepository)(nil)
