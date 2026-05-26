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
	acct, err := s.accountRepo.FindByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, identity.ErrNotFound
	}
	return newAccountDTO(acct, u), nil
}
