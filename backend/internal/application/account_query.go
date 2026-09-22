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
	if u.Subject == "" {
		return nil, auth.ErrResHubSession
	}
	return newExternalAccountDTO(u), nil
}

func newExternalAccountDTO(u *auth.User) *AccountDTO {
	return &AccountDTO{
		ID:               u.ID,
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
