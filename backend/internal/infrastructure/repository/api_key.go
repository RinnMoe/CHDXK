package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/auth"
)

type ApiKeyRepository struct {
	db *gorm.DB
}

func NewApiKeyRepository(db *gorm.DB) *ApiKeyRepository {
	return &ApiKeyRepository{db: db}
}

func newApiKeyDomain(e *ApiKeyEntity) auth.ApiKey {
	userID := 0
	if e.UserID != nil {
		userID = *e.UserID
	}
	return auth.ApiKey{
		ID:         e.ID,
		Name:       e.Name,
		Key:        e.Key,
		Role:       e.Role,
		UserID:     userID,
		CreatedAt:  e.CreatedAt,
		LastUsedAt: e.LastUsedAt,
	}
}

func newApiKeyEntity(d *auth.ApiKey) ApiKeyEntity {
	var userID *int
	if d.UserID > 0 {
		userID = &d.UserID
	}
	return ApiKeyEntity{
		ID:         d.ID,
		Name:       d.Name,
		Key:        d.Key,
		Role:       d.Role,
		UserID:     userID,
		CreatedAt:  d.CreatedAt,
		LastUsedAt: d.LastUsedAt,
	}
}

func (r *ApiKeyRepository) FindByKey(ctx context.Context, key string) (*auth.ApiKey, error) {
	var e ApiKeyEntity
	if err := r.db.WithContext(ctx).
		Select("id, name, key, role, user_id, last_used_at, created_at").
		Where("key = ?", key).
		Take(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := newApiKeyDomain(&e)
	return &d, nil
}

func (r *ApiKeyRepository) ListByUser(ctx context.Context, userID int) ([]auth.ApiKey, error) {
	var entities []ApiKeyEntity
	if err := r.db.WithContext(ctx).
		Where("role = ? AND user_id = ?", auth.ApiKeyRoleUser, userID).
		Order("created_at DESC, id DESC").
		Find(&entities).Error; err != nil {
		return nil, err
	}
	keys := make([]auth.ApiKey, 0, len(entities))
	for i := range entities {
		keys = append(keys, newApiKeyDomain(&entities[i]))
	}
	return keys, nil
}

func (r *ApiKeyRepository) Create(ctx context.Context, apiKey *auth.ApiKey) error {
	e := newApiKeyEntity(apiKey)
	if err := r.db.WithContext(ctx).Create(&e).Error; err != nil {
		return err
	}
	apiKey.ID = e.ID
	apiKey.CreatedAt = e.CreatedAt
	return nil
}

func (r *ApiKeyRepository) DeleteByUser(ctx context.Context, id int, userID int) (bool, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND role = ? AND user_id = ?", id, auth.ApiKeyRoleUser, userID).
		Delete(&ApiKeyEntity{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *ApiKeyRepository) TouchLastUsed(ctx context.Context, id int, at time.Time) error {
	return r.db.WithContext(ctx).
		Model(&ApiKeyEntity{}).
		Where("id = ?", id).
		Update("last_used_at", at).Error
}

var _ auth.ApiKeyRepository = (*ApiKeyRepository)(nil)
