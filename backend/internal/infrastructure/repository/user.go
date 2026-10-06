package repository

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"jcourse/internal/domain/auth"
)

type UserRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func newUserDomain(e *UserEntity) auth.User {
	u := auth.User{ID: e.ID, Role: e.Role, LocalRole: e.Role, Username: e.Username, Nickname: e.Nickname, Email: e.Email.String, IsBootstrap: e.ID == 1, MustChangePassword: e.MustChangePassword, HasPassword: e.PasswordHash != "", AuthVersion: e.AuthVersion, MutedUntil: e.MutedUntil, ModerationReason: e.ModerationReason, DeletedAt: e.DeletedAt, SuspendedAt: e.SuspendedAt, SuspendTill: e.SuspendTill}
	_ = json.Unmarshal([]byte(e.Permissions), &u.Permissions)
	if u.Permissions == nil {
		u.Permissions = []string{}
	}
	return u
}

func (r *UserRepository) Update(ctx context.Context, u *auth.User) error {
	if err := r.db.WithContext(ctx).
		Model(&UserEntity{}).
		Where("id = ?", u.ID).
		Updates(map[string]any{
			"suspended_at": u.SuspendedAt,
			"suspend_till": u.SuspendTill,
		}).Error; err != nil {
		return err
	}
	cacheDelete(ctx, r.cache, cacheKey("user", u.ID))
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (*auth.User, error) {
	return loadAccount(r.db.WithContext(ctx), id)
}

func NewUserRepository(db *gorm.DB, cache ...*redis.Client) *UserRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &UserRepository{db: db, cache: client}
}

var _ auth.UserRepository = (*UserRepository)(nil)
