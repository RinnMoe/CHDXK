package application

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"jcourse/internal/domain/account/credential"
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
	adminUsers  *auth.AdminUserService
	accountRepo identity.Repository
	hasher      credential.PasswordHasher
	settings    SiteSettingsProvider
}

type AdminUserCommandConfig = auth.AdminConfig

var DefaultAdminUserCommandConfig = AdminUserCommandConfig{DefaultSuspendDays: auth.DefaultAdminConfig.DefaultSuspendDays}

func NewAdminUserCommandService(
	userRepo auth.UserRepository,
	accountRepo identity.Repository,
	hasher credential.PasswordHasher,
	settings SiteSettingsProvider,
) *AdminUserCommandService {
	if settings == nil {
		defaults := NewDefaultSiteSettingsProvider()
		settings = defaults
	}
	return &AdminUserCommandService{
		adminUsers:  auth.NewAdminUserService(userRepo, auth.AdminConfig{}),
		accountRepo: accountRepo,
		hasher:      hasher,
		settings:    settings,
	}
}

func (s *AdminUserCommandService) SuspendUserForDays(ctx context.Context, actor *auth.User, userID int, days int) error {
	if days <= 0 {
		config, err := s.settings.AdminUserConfig(ctx)
		if err != nil {
			return err
		}
		days = config.DefaultSuspendDays
	}
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

func (s *AdminUserCommandService) ResetPassword(ctx context.Context, actor *auth.User, userID int, newPassword string) error {
	if strings.TrimSpace(newPassword) == "" {
		return credential.ErrPasswordRequired
	}

	acct, err := s.accountRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if acct == nil {
		return identity.ErrNotFound
	}

	passwordHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	hadPassword := acct.PasswordHash != ""
	acct.PasswordHash = passwordHash
	if err := s.accountRepo.Update(ctx, acct); err != nil {
		return err
	}

	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  time.Now(),
		ActorUserID: actor.ID,
		Action:      audit.ActionUserPasswordReset,
		TargetType:  audit.TargetTypeUser,
		TargetID:    strconv.Itoa(userID),
		Details: audit.Details{
			"email":               acct.Email,
			"had_password_before": hadPassword,
		},
	})
	return nil
}

func mapAuthUserNotFound(err error) error {
	if errors.Is(err, auth.ErrUserNotFound) {
		return identity.ErrNotFound
	}
	return err
}
