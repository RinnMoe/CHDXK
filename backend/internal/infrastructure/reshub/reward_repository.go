package reshub

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
	"jcourse/internal/infrastructure/repository"
)

type RewardRepository struct {
	db         *gorm.DB
	provider   point.ExternalPointProvider
	identities auth.IdentityProjectionRepository
}

var errRemoteOperationNotFound = errors.New("res-hub 积分操作不存在")

func NewRewardRepository(db *gorm.DB, provider point.ExternalPointProvider, identities auth.IdentityProjectionRepository) *RewardRepository {
	return &RewardRepository{db: db, provider: provider, identities: identities}
}

func (r *RewardRepository) GrantReward(ctx context.Context, rewardID int, now time.Time) error {
	var entity repository.PointRewardEntity
	if err := r.db.WithContext(ctx).Where("id = ?", rewardID).Take(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if entity.Status == string(point.RewardStatusGranted) {
		return nil
	}
	if entity.Status == string(point.RewardStatusCanceled) {
		if entity.ExternalReversalID == "" && entity.OperationKey != "" {
			return r.reconcileCanceled(ctx, entity)
		}
		return nil
	}

	subject, err := r.identities.SubjectByUserID(ctx, entity.UserID)
	if err != nil {
		return r.recordFailure(ctx, rewardID, err)
	}
	operationKey := entity.OperationKey
	if operationKey == "" {
		operationKey = fmt.Sprintf("xk:point-reward:%d", rewardID)
	}
	if err := r.persistOperationKey(ctx, rewardID, operationKey); err != nil {
		return err
	}
	operation, err := r.provider.Change(ctx, subject, entity.Amount, entity.Reason, operationKey)
	if err != nil {
		return r.recordFailure(ctx, rewardID, err)
	}
	updates := map[string]any{
		"status":                string(point.RewardStatusGranted),
		"granted_at":            now,
		"operation_key":         operationKey,
		"external_operation_id": operation.ID,
		"last_error":            "",
		"next_retry_at":         nil,
	}
	result := r.db.WithContext(ctx).Model(&repository.PointRewardEntity{}).
		Where("id = ? AND status = ?", rewardID, string(point.RewardStatusPending)).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 1 {
		return nil
	}

	// Revoke may have canceled the local reward while the remote change was
	// in flight. Compensate the idempotent operation before returning success.
	var current repository.PointRewardEntity
	if err := r.db.WithContext(ctx).Where("id = ?", rewardID).Take(&current).Error; err != nil {
		return err
	}
	if current.Status != string(point.RewardStatusCanceled) || current.ExternalReversalID != "" {
		return nil
	}
	if err := r.db.WithContext(ctx).Model(&repository.PointRewardEntity{}).Where("id = ?", rewardID).
		Update("external_operation_id", operation.ID).Error; err != nil {
		return err
	}
	current.OperationKey = operationKey
	current.ExternalOperationID = operation.ID
	return r.reverseCanceled(ctx, current, subject, operation.ID)
}

func (r *RewardRepository) RevokeRewardsBySources(ctx context.Context, sources []point.RewardSource, now time.Time) error {
	for _, source := range sources {
		var entities []repository.PointRewardEntity
		if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND reason = ? AND source_type = ? AND source_key = ?", source.UserID, string(source.Reason), source.SourceType, source.SourceKey).
			Find(&entities).Error; err != nil {
			return err
		}
		for _, entity := range entities {
			switch entity.Status {
			case string(point.RewardStatusCanceled):
				continue
			case string(point.RewardStatusPending):
				result := r.db.WithContext(ctx).Model(&repository.PointRewardEntity{}).
					Where("id = ? AND status = ?", entity.ID, string(point.RewardStatusPending)).
					Updates(map[string]any{"status": string(point.RewardStatusCanceled), "last_error": "", "next_retry_at": nil})
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					var current repository.PointRewardEntity
					if err := r.db.WithContext(ctx).Where("id = ?", entity.ID).Take(&current).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							continue
						}
						return err
					}
					if current.Status == string(point.RewardStatusGranted) {
						if err := r.revokeGranted(ctx, current); err != nil {
							return err
						}
					}
				}
			case string(point.RewardStatusGranted):
				if err := r.revokeGranted(ctx, entity); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *RewardRepository) revokeGranted(ctx context.Context, entity repository.PointRewardEntity) error {
	subject, err := r.identities.SubjectByUserID(ctx, entity.UserID)
	if err != nil {
		return r.recordFailure(ctx, entity.ID, err)
	}
	operationID, err := r.resolveOperationID(ctx, entity, subject)
	if err != nil {
		return r.recordFailure(ctx, entity.ID, err)
	}
	if operationID == "" {
		return r.recordFailure(ctx, entity.ID, errRemoteOperationNotFound)
	}
	reversal, err := r.provider.Reverse(ctx, subject, operationID)
	if err != nil {
		return r.recordFailure(ctx, entity.ID, err)
	}
	return r.db.WithContext(ctx).Model(&repository.PointRewardEntity{}).Where("id = ?", entity.ID).Updates(map[string]any{
		"status":                string(point.RewardStatusCanceled),
		"external_operation_id": operationID,
		"external_reversal_id":  reversal.ID,
		"last_error":            "",
		"next_retry_at":         nil,
		"granted_at":            entity.GrantedAt,
	}).Error
}

func (r *RewardRepository) persistOperationKey(ctx context.Context, rewardID int, operationKey string) error {
	return r.db.WithContext(ctx).Model(&repository.PointRewardEntity{}).
		Where("id = ? AND (operation_key IS NULL OR operation_key = '')", rewardID).
		Update("operation_key", operationKey).Error
}

func (r *RewardRepository) reconcileCanceled(ctx context.Context, entity repository.PointRewardEntity) error {
	subject, err := r.identities.SubjectByUserID(ctx, entity.UserID)
	if err != nil {
		return r.recordFailure(ctx, entity.ID, err)
	}
	operation, err := r.provider.FindOperation(ctx, subject, entity.OperationKey)
	if err != nil {
		return r.recordFailure(ctx, entity.ID, err)
	}
	if operation == nil {
		return nil
	}
	if strings.TrimSpace(operation.ID) == "" {
		return nil
	}
	if entity.ExternalOperationID == "" {
		if err := r.db.WithContext(ctx).Model(&repository.PointRewardEntity{}).Where("id = ?", entity.ID).
			Update("external_operation_id", operation.ID).Error; err != nil {
			return err
		}
		entity.ExternalOperationID = operation.ID
	}
	return r.reverseCanceled(ctx, entity, subject, operation.ID)
}

func (r *RewardRepository) reverseCanceled(ctx context.Context, entity repository.PointRewardEntity, subject, operationID string) error {
	if entity.ExternalReversalID != "" {
		return nil
	}
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return r.recordFailure(ctx, entity.ID, errRemoteOperationNotFound)
	}
	reversal, err := r.provider.Reverse(ctx, subject, operationID)
	if err != nil {
		return r.recordFailure(ctx, entity.ID, err)
	}
	updates := map[string]any{
		"external_operation_id": operationID,
		"external_reversal_id":  reversal.ID,
		"last_error":            "",
		"next_retry_at":         nil,
	}
	return r.db.WithContext(ctx).Model(&repository.PointRewardEntity{}).
		Where("id = ? AND status = ?", entity.ID, string(point.RewardStatusCanceled)).
		Updates(updates).Error
}

func (r *RewardRepository) resolveOperationID(ctx context.Context, entity repository.PointRewardEntity, subject string) (string, error) {
	if operationID := strings.TrimSpace(entity.ExternalOperationID); operationID != "" {
		return operationID, nil
	}
	operationKey := strings.TrimSpace(entity.OperationKey)
	if operationKey == "" {
		operationKey = fmt.Sprintf("xk:point-reward:%d", entity.ID)
	}
	operation, err := r.provider.FindOperation(ctx, subject, operationKey)
	if err != nil {
		return "", err
	}
	if operation == nil || strings.TrimSpace(operation.ID) == "" {
		return "", nil
	}
	operationID := strings.TrimSpace(operation.ID)
	if err := r.db.WithContext(ctx).Model(&repository.PointRewardEntity{}).Where("id = ?", entity.ID).
		Update("external_operation_id", operationID).Error; err != nil {
		return "", err
	}
	return operationID, nil
}

func (r *RewardRepository) recordFailure(ctx context.Context, rewardID int, err error) error {
	message := strings.TrimSpace(err.Error())
	if len(message) > 1000 {
		message = message[:1000]
	}
	returnErr := err
	if updateErr := r.db.WithContext(ctx).Model(&repository.PointRewardEntity{}).Where("id = ?", rewardID).
		Updates(map[string]any{
			"attempts":      gorm.Expr("attempts + 1"),
			"last_error":    message,
			"next_retry_at": time.Now().Add(time.Minute),
		}).Error; updateErr != nil {
		returnErr = errors.Join(returnErr, updateErr)
	}
	return returnErr
}

var _ point.RewardRepository = (*RewardRepository)(nil)
