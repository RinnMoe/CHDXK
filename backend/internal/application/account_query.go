package application

import (
	"context"

	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
)

type AccountQueryService struct {
	accountRepo identity.Repository
}

func NewAccountQueryService(accountRepo identity.Repository) *AccountQueryService {
	return &AccountQueryService{accountRepo: accountRepo}
}

func (s *AccountQueryService) CurrentUser(ctx context.Context, u *auth.User) (*AccountDTO, error) {
	if u == nil {
		return nil, identity.ErrNotFound
	}
	if u.Subject != "" {
		return newExternalAccountDTO(u), nil
	}
	acct, err := s.accountRepo.FindByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, identity.ErrNotFound
	}
	return newAccountDTO(acct, u), nil
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
