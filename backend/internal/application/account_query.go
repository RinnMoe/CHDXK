package application

import (
	"context"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
)

type AccountQueryService struct {
	accountRepo account.AccountRepository
}

func NewAccountQueryService(accountRepo account.AccountRepository) *AccountQueryService {
	return &AccountQueryService{accountRepo: accountRepo}
}

func (s *AccountQueryService) CurrentUser(ctx context.Context, u *auth.User) (*AccountDTO, error) {
	acct, err := s.accountRepo.FindByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, account.ErrUserNotFound
	}
	return newAccountDTO(acct, u), nil
}
