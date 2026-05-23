package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
)

type AccountRepository struct {
	db *gorm.DB
}

type UserRepository struct {
	db *gorm.DB
}

func newAccountEntity(u *account.Account) UserEntity {
	return UserEntity{
		ID:           u.ID,
		Username:     u.Username,
		Role:         auth.RoleUser,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		LastSeenAt:   u.LastSeenAt,
	}
}

func newAccountDomain(e *UserEntity) account.Account {
	return account.Account{
		ID:           e.ID,
		Username:     e.Username,
		Email:        e.Email,
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

func (r *AccountRepository) Create(ctx context.Context, u *account.Account) error {
	e := newAccountEntity(u)
	if err := gorm.G[UserEntity](r.db).Create(ctx, &e); err != nil {
		return err
	}
	u.ID = e.ID
	return nil
}

func (r *AccountRepository) Update(ctx context.Context, u *account.Account) error {
	return r.db.WithContext(ctx).
		Model(&UserEntity{}).
		Where("id = ?", u.ID).
		Updates(map[string]any{
			"username":      u.Username,
			"email":         u.Email,
			"password_hash": u.PasswordHash,
			"last_seen_at":  u.LastSeenAt,
		}).Error
}

func (r *AccountRepository) TouchLastSeen(ctx context.Context, userID int, at time.Time) error {
	return r.db.WithContext(ctx).
		Model(&UserEntity{}).
		Where("id = ?", userID).
		Update("last_seen_at", at).Error
}

func (r *AccountRepository) FindByID(ctx context.Context, id int) (*account.Account, error) {
	e, err := gorm.G[UserEntity](r.db).Where("id = ?", id).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := newAccountDomain(&e)
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
	e, err := gorm.G[UserEntity](r.db).Where("email = ?", email).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := newAccountDomain(&e)
	return &d, nil
}

func (r *UserRepository) Update(ctx context.Context, u *auth.User) error {
	return r.db.WithContext(ctx).
		Model(&UserEntity{}).
		Where("id = ?", u.ID).
		Updates(map[string]any{
			"role":         u.Role,
			"suspended_at": u.SuspendedAt,
			"suspend_till": u.SuspendTill,
		}).Error
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (*auth.User, error) {
	e, err := gorm.G[UserEntity](r.db).Where("id = ?", id).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := newUserDomain(&e)
	return &d, nil
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

var _ account.AccountRepository = (*AccountRepository)(nil)
var _ auth.UserRepository = (*UserRepository)(nil)
