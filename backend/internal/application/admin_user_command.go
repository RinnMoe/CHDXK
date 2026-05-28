package application

import (
	"context"
	"errors"
	"strconv"
	"time"

	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
)

var (
	ErrCannotSuspendAdmin     = auth.ErrCannotSuspendAdmin
	ErrCannotOperateSelf      = auth.ErrCannotOperateSelf
	ErrCannotModifySuperAdmin = auth.ErrCannotModifySuperAdmin
)

type AdminUserCommandService struct {
	adminUsers *auth.AdminUserService
}

type AdminUserCommandConfig = auth.AdminConfig

var DefaultAdminUserCommandConfig = AdminUserCommandConfig{DefaultSuspendDays: auth.DefaultAdminConfig.DefaultSuspendDays}

func NewAdminUserCommandService(userRepo auth.UserRepository, config AdminUserCommandConfig) *AdminUserCommandService {
	return &AdminUserCommandService{adminUsers: auth.NewAdminUserService(userRepo, auth.AdminConfig(config))}
}

func (s *AdminUserCommandService) SuspendUserForDays(ctx context.Context, actor *auth.User, userID int, days int) error {
	err := mapAuthUserNotFound(s.adminUsers.SuspendUserForDays(ctx, actor.ID, userID, days))
	if err != nil {
		return err
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  time.Now(),
		ActorUserID: actor.ID,
		Action:      audit.ActionUserSuspend,
		TargetType:  audit.TargetTypeUser,
		TargetID:    strconv.Itoa(userID),
		Details: audit.Details{
			"days": days,
		},
	})
	return nil
}

func (s *AdminUserCommandService) ClearSuspension(ctx context.Context, actor *auth.User, userID int) error {
	err := mapAuthUserNotFound(s.adminUsers.ClearSuspension(ctx, actor.ID, userID))
	if err != nil {
		return err
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  time.Now(),
		ActorUserID: actor.ID,
		Action:      audit.ActionUserUnsuspend,
		TargetType:  audit.TargetTypeUser,
		TargetID:    strconv.Itoa(userID),
	})
	return nil
}

func (s *AdminUserCommandService) GrantAdmin(ctx context.Context, actor *auth.User, userID int) error {
	err := mapAuthUserNotFound(s.adminUsers.GrantAdmin(ctx, actor.ID, userID))
	if err != nil {
		return err
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  time.Now(),
		ActorUserID: actor.ID,
		Action:      audit.ActionAdminGrant,
		TargetType:  audit.TargetTypeUser,
		TargetID:    strconv.Itoa(userID),
	})
	return nil
}

func (s *AdminUserCommandService) RevokeAdmin(ctx context.Context, actor *auth.User, userID int) error {
	err := mapAuthUserNotFound(s.adminUsers.RevokeAdmin(ctx, actor.ID, userID))
	if err != nil {
		return err
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  time.Now(),
		ActorUserID: actor.ID,
		Action:      audit.ActionAdminRevoke,
		TargetType:  audit.TargetTypeUser,
		TargetID:    strconv.Itoa(userID),
	})
	return nil
}

func mapAuthUserNotFound(err error) error {
	if errors.Is(err, auth.ErrUserNotFound) {
		return identity.ErrNotFound
	}
	return err
}
