package auth

import (
	"context"
	"time"

	"jcourse/pkg/apperr"
)

var (
	ErrCannotSuspendAdmin     = apperr.ErrCannotSuspendAdmin
	ErrCannotOperateSelf      = apperr.ErrCannotOperateSelf
	ErrCannotModifySuperAdmin = apperr.ErrCannotModifySuperAdmin
	ErrUserNotFound           = apperr.ErrUserNotFound
)

type AdminUserService struct {
	userRepo UserRepository
	config   AdminConfig
}

func NewAdminUserService(userRepo UserRepository, config AdminConfig) *AdminUserService {
	if config.DefaultSuspendDays <= 0 {
		config.DefaultSuspendDays = DefaultAdminConfig.DefaultSuspendDays
	}
	return &AdminUserService{userRepo: userRepo, config: config}
}

func (s *AdminUserService) SuspendUserForDays(ctx context.Context, actorUserID int, userID int, days int) error {
	if actorUserID == userID {
		return ErrCannotOperateSelf
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrUserNotFound
	}
	if u.IsAdmin() {
		return ErrCannotSuspendAdmin
	}
	if days <= 0 {
		days = s.config.DefaultSuspendDays
	}

	u.Suspend(time.Duration(days) * 24 * time.Hour)
	return s.userRepo.Update(ctx, u)
}

func (s *AdminUserService) ClearSuspension(ctx context.Context, actorUserID int, userID int) error {
	if actorUserID == userID {
		return ErrCannotOperateSelf
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrUserNotFound
	}

	u.ClearSuspension()
	return s.userRepo.Update(ctx, u)
}

func (s *AdminUserService) GrantAdmin(ctx context.Context, actorUserID int, userID int) error {
	if actorUserID == userID {
		return ErrCannotOperateSelf
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrUserNotFound
	}
	if u.IsSuperAdmin() {
		return ErrCannotModifySuperAdmin
	}

	u.Role = RoleAdmin
	u.ClearSuspension()
	return s.userRepo.Update(ctx, u)
}

func (s *AdminUserService) RevokeAdmin(ctx context.Context, actorUserID int, userID int) error {
	if actorUserID == userID {
		return ErrCannotOperateSelf
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrUserNotFound
	}
	if u.IsSuperAdmin() {
		return ErrCannotModifySuperAdmin
	}

	u.Role = RoleUser
	return s.userRepo.Update(ctx, u)
}
