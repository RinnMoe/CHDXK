package repository

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"jcourse/internal/domain/auth"
)

type UserRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func newUserDomain(e *UserEntity) auth.User {
	return auth.User{
		ID:          e.ID,
		Role:        e.Role,
		SuspendedAt: e.SuspendedAt,
		SuspendTill: e.SuspendTill,
	}
}

func (r *UserRepository) Update(ctx context.Context, u *auth.User) error {
	if err := r.db.WithContext(ctx).
		Model(&UserEntity{}).
		Where("id = ?", u.ID).
		Updates(map[string]any{
			"role":         u.Role,
			"suspended_at": u.SuspendedAt,
			"suspend_till": u.SuspendTill,
		}).Error; err != nil {
		return err
	}
	cacheDelete(ctx, r.cache, cacheKey("user", u.ID))
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (*auth.User, error) {
	key := cacheKey("user", id)
	if cached, ok := cacheGetJSON[auth.User](ctx, r.cache, key); ok {
		return cached, nil
	}

	e, err := gorm.G[UserEntity](r.db).Where("id = ?", id).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := newUserDomain(&e)
	cacheSetJSON(ctx, r.cache, key, &d)
	return &d, nil
}

func NewUserRepository(db *gorm.DB, cache ...*redis.Client) *UserRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &UserRepository{db: db, cache: client}
}

var _ auth.UserRepository = (*UserRepository)(nil)
