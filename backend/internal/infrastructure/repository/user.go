package repository

import (
	"context"

	"gorm.io/gorm"

	"jcourse/internal/domain/auth"
)

type UserRepository struct {
	db *gorm.DB
}

func newUserEntity(u *auth.User) UserEntity {
	return UserEntity{
		ID:          u.ID,
		Username:    u.Username,
		Role:        u.Role,
		Email:       u.Email,
		Password:    u.Password,
		CreatedAt:   u.CreatedAt,
		LastSeenAt:  u.LastSeenAt,
		SuspendedAt: u.SuspendedAt,
		SuspendTill: u.SuspendTill,
	}
}

func newUserDomain(e *UserEntity) auth.User {
	return auth.User{
		ID:          e.ID,
		Username:    e.Username,
		Role:        e.Role,
		Email:       e.Email,
		Password:    e.Password,
		CreatedAt:   e.CreatedAt,
		LastSeenAt:  e.LastSeenAt,
		SuspendedAt: e.SuspendedAt,
	}
}

func (u2 *UserRepository) Create(ctx context.Context, u *auth.User) error {
	e := newUserEntity(u)
	if err := gorm.G[UserEntity](u2.db).Create(ctx, &e); err != nil {
		return err
	}
	u.ID = e.ID
	return nil
}

func (u2 *UserRepository) Update(ctx context.Context, u *auth.User) error {
	e := newUserEntity(u)
	if _, err := gorm.G[UserEntity](u2.db).Updates(ctx, e); err != nil {
		return err
	}
	return nil
}

func (u2 *UserRepository) FindByID(ctx context.Context, id int) (*auth.User, error) {
	e, err := gorm.G[UserEntity](u2.db).Where("id = ?", id).Take(ctx)
	if err != nil {
		return nil, err
	}
	return new(newUserDomain(&e)), nil
}

func (u2 *UserRepository) FindByUsername(ctx context.Context, username string) (*auth.User, error) {
	e, err := gorm.G[UserEntity](u2.db).Where("username = ?", username).Take(ctx)
	if err != nil {
		return nil, err
	}
	return new(newUserDomain(&e)), nil
}

func (u2 *UserRepository) FindByEmail(ctx context.Context, email string) (*auth.User, error) {
	e, err := gorm.G[UserEntity](u2.db).Where("email = ?", email).Take(ctx)
	if err != nil {
		return nil, err
	}
	return new(newUserDomain(&e)), nil
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

var _ auth.UserRepository = (*UserRepository)(nil)
