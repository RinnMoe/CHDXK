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
	return auth.ApiKey{
		ID:        e.ID,
		Name:      e.Name,
		Key:       e.Key,
		CreatedAt: e.CreatedAt,
	}
}

func (r *ApiKeyRepository) FindByKey(ctx context.Context, key string) (*auth.ApiKey, error) {
	var e ApiKeyEntity
	if err := r.db.WithContext(ctx).Where("`key` = ?", key).Take(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := newApiKeyDomain(&e)
	return &d, nil
}

var _ auth.ApiKeyRepository = (*ApiKeyRepository)(nil)
