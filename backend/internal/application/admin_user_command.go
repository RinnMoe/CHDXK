package application

import (
	"context"
	"errors"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
)

var (
	ErrCannotSuspendAdmin = auth.ErrCannotSuspendAdmin
	ErrCannotOperateSelf  = auth.ErrCannotOperateSelf
)

type AdminUserCommandService struct {
	adminUsers *auth.AdminUserService
}

type AdminUserCommandConfig = auth.AdminConfig

var DefaultAdminUserCommandConfig = AdminUserCommandConfig{DefaultSuspendDays: auth.DefaultAdminConfig.DefaultSuspendDays}

func NewAdminUserCommandService(userRepo auth.UserRepository, config AdminUserCommandConfig) *AdminUserCommandService {
	return &AdminUserCommandService{adminUsers: auth.NewAdminUserService(userRepo, auth.AdminConfig(config))}
}

func (s *AdminUserCommandService) SuspendUserForDays(ctx context.Context, actorUserID int, userID int, days int) error {
	return mapAuthUserNotFound(s.adminUsers.SuspendUserForDays(ctx, actorUserID, userID, days))
}

func (s *AdminUserCommandService) ClearSuspension(ctx context.Context, actorUserID int, userID int) error {
	return mapAuthUserNotFound(s.adminUsers.ClearSuspension(ctx, actorUserID, userID))
}

func (s *AdminUserCommandService) GrantAdmin(ctx context.Context, actorUserID int, userID int) error {
	return mapAuthUserNotFound(s.adminUsers.GrantAdmin(ctx, actorUserID, userID))
}

func (s *AdminUserCommandService) RevokeAdmin(ctx context.Context, actorUserID int, userID int) error {
	return mapAuthUserNotFound(s.adminUsers.RevokeAdmin(ctx, actorUserID, userID))
}

func mapAuthUserNotFound(err error) error {
	if errors.Is(err, auth.ErrUserNotFound) {
		return account.ErrUserNotFound
	}
	return err
}
