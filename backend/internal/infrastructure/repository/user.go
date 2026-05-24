package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
)

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

type AccountRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

type UserRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func newAccountEntity(u *account.Account) UserEntity {
	return UserEntity{
		ID:           u.ID,
		Username:     u.Username,
		Role:         auth.RoleUser,
		Email:        nullString(u.Email),
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		LastSeenAt:   u.LastSeenAt,
	}
}

func newAccountDomain(e *UserEntity) account.Account {
	return account.Account{
		ID:           e.ID,
		Username:     e.Username,
		Email:        e.Email.String,
		PasswordHash: e.PasswordHash,
		CreatedAt:    e.CreatedAt,
		LastSeenAt:   e.LastSeenAt,
	}
}

func newUserDomain(e *UserEntity) auth.User {
	return auth.User{
		ID:          e.ID,
		Role:        e.Role,
		SuspendedAt: e.SuspendedAt,
		SuspendTill: e.SuspendTill,
	}
}

func (r *AccountRepository) deleteAccountCache(ctx context.Context, userID int, emails ...string) {
	keys := []string{cacheKey("account", userID)}
	for _, email := range emails {
		if email != "" {
			keys = append(keys, cacheKey("account", "email", email))
		}
	}
	cacheDelete(ctx, r.cache, keys...)
}

func (r *AccountRepository) cachedEmailByID(ctx context.Context, userID int) string {
	var e UserEntity
	if err := r.db.WithContext(ctx).Select("email").Where("id = ?", userID).Take(&e).Error; err != nil {
		return ""
	}
	return e.Email.String
}

func (r *AccountRepository) Create(ctx context.Context, u *account.Account) error {
	e := newAccountEntity(u)
	if err := gorm.G[UserEntity](r.db).Create(ctx, &e); err != nil {
		return err
	}
	u.ID = e.ID
	r.deleteAccountCache(ctx, u.ID, u.Email)
	return nil
}

func (r *AccountRepository) Update(ctx context.Context, u *account.Account) error {
	oldEmail := r.cachedEmailByID(ctx, u.ID)
	if err := r.db.WithContext(ctx).
		Model(&UserEntity{}).
		Where("id = ?", u.ID).
		Updates(map[string]any{
			"username":      u.Username,
			"email":         nullString(u.Email),
			"password_hash": u.PasswordHash,
			"last_seen_at":  u.LastSeenAt,
		}).Error; err != nil {
		return err
	}
	r.deleteAccountCache(ctx, u.ID, oldEmail, u.Email)
	return nil
}

func (r *AccountRepository) TouchLastSeen(ctx context.Context, userID int, at time.Time) error {
	email := r.cachedEmailByID(ctx, userID)
	if err := r.db.WithContext(ctx).
		Model(&UserEntity{}).
		Where("id = ?", userID).
		Update("last_seen_at", at).Error; err != nil {
		return err
	}
	r.deleteAccountCache(ctx, userID, email)
	return nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id int) (*account.Account, error) {
	key := cacheKey("account", id)
	if cached, ok := cacheGetJSON[account.Account](ctx, r.cache, key); ok {
		return cached, nil
	}

	e, err := gorm.G[UserEntity](r.db).Where("id = ?", id).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := newAccountDomain(&e)
	cacheSetJSON(ctx, r.cache, key, &d)
	return &d, nil
}

func (r *AccountRepository) FindByUsername(ctx context.Context, username string) (*account.Account, error) {
	e, err := gorm.G[UserEntity](r.db).Where("username = ?", username).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := newAccountDomain(&e)
	return &d, nil
}

func (r *AccountRepository) FindByEmail(ctx context.Context, email string) (*account.Account, error) {
	key := cacheKey("account", "email", email)
	if cached, ok := cacheGetJSON[account.Account](ctx, r.cache, key); ok {
		return cached, nil
	}

	e, err := gorm.G[UserEntity](r.db).Where("email = ?", email).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := newAccountDomain(&e)
	cacheSetJSON(ctx, r.cache, key, &d)
	return &d, nil
}

func (r *UserRepository) Update(ctx context.Context, u *auth.User) error {
	var e UserEntity
	email := ""
	if err := r.db.WithContext(ctx).Select("email").Where("id = ?", u.ID).Take(&e).Error; err == nil {
		email = e.Email.String
	}
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
	keys := []string{cacheKey("user", u.ID), cacheKey("account", u.ID)}
	if email != "" {
		keys = append(keys, cacheKey("account", "email", email))
	}
	cacheDelete(ctx, r.cache, keys...)
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

func (r *UserRepository) FindByRole(ctx context.Context, role string) ([]auth.User, error) {
	es, err := gorm.G[UserEntity](r.db).Where("role = ?", role).Order("id ASC").Find(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]auth.User, len(es))
	for i := range es {
		users[i] = newUserDomain(&es[i])
	}
	return users, nil
}

func NewAccountRepository(db *gorm.DB, cache ...*redis.Client) *AccountRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &AccountRepository{db: db, cache: client}
}

func NewUserRepository(db *gorm.DB, cache ...*redis.Client) *UserRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &UserRepository{db: db, cache: client}
}

var _ account.AccountRepository = (*AccountRepository)(nil)
var _ auth.UserRepository = (*UserRepository)(nil)
