package application

import (
	"context"

	"jcourse/internal/domain/auth"
)

type AccountQueryService struct{}

func NewAccountQueryService() *AccountQueryService {
	return &AccountQueryService{}
}

func (s *AccountQueryService) CurrentUser(_ context.Context, u *auth.User) (*AccountDTO, error) {
	if u == nil {
		return nil, auth.ErrResHubSession
	}
	if u.ID <= 0 {
		return nil, auth.ErrResHubSession
	}
	return newExternalAccountDTO(u), nil
}

func NewAccountDTO(u *auth.User) *AccountDTO { return newExternalAccountDTO(u) }
func newExternalAccountDTO(u *auth.User) *AccountDTO {
	return &AccountDTO{
		ID: u.ID, Nickname: u.Nickname, LocalRole: u.LocalRole, Permissions: u.Permissions, MustChangePassword: u.MustChangePassword, HasPassword: u.HasPassword, SuspendedAt: u.SuspendedAt, SuspendTill: u.SuspendTill, MutedUntil: u.MutedUntil, ModerationReason: u.ModerationReason,
		IsBootstrap:      u.IsBootstrap,
		BoundNickname:    u.BoundNickname,
		Username:         u.Username,
		Email:            u.Email,
		Role:             u.Role,
		Subject:          u.Subject,
		IdentityRole:     u.IdentityRole,
		IdentityType:     u.IdentityType,
		PermissionMode:   u.PermissionMode,
		ViolationCount:   u.ViolationCount,
		InitialPassword:  u.InitialPassword,
		HasPointsAccount: u.HasPointsAccount,
		SessionID:        u.SessionID,
		ExpiresAt:        u.ExpiresAt,
		ImpersonatedBy:   u.ImpersonatedBy,
	}
}
