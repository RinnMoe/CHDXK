package repository

import (
	"context"
	"errors"

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
		SecretHash: e.SecretHash,
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
		SecretHash: d.SecretHash,
		Role:       d.Role,
		UserID:     userID,
		CreatedAt:  d.CreatedAt,
		LastUsedAt: d.LastUsedAt,
	}
}

func (r *ApiKeyRepository) GetByID(ctx context.Context, id int64) (*auth.ApiKey, error) {
	var e ApiKeyEntity
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Take(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return new(newApiKeyDomain(&e)), nil
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

func (r *ApiKeyRepository) ListSystem(ctx context.Context) ([]auth.ApiKey, error) {
	var entities []ApiKeyEntity
	if err := r.db.WithContext(ctx).
		Where("role = ? AND user_id IS NULL", auth.ApiKeyRoleSystem).
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

func (r *ApiKeyRepository) CountByUser(ctx context.Context, userID int) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&ApiKeyEntity{}).
		Where("role = ? AND user_id = ?", auth.ApiKeyRoleUser, userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
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

func (r *ApiKeyRepository) Delete(ctx context.Context, id int64) (bool, error) {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&ApiKeyEntity{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

var _ auth.ApiKeyRepository = (*ApiKeyRepository)(nil)
var _ auth.ApiKeyQuery = (*ApiKeyRepository)(nil)
