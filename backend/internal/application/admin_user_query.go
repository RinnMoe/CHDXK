package application

import (
	"context"
	"strings"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
)

type AdminUserQueryService struct {
	accountRepo account.AccountRepository
	userRepo    auth.UserRepository
	usernames   account.UsernameDeriver
}

func NewAdminUserQueryService(
	accountRepo account.AccountRepository,
	userRepo auth.UserRepository,
	usernames account.UsernameDeriver,
) *AdminUserQueryService {
	return &AdminUserQueryService{accountRepo: accountRepo, userRepo: userRepo, usernames: usernames}
}

func (s *AdminUserQueryService) FindByEmail(ctx context.Context, email string) (*AdminUserDTO, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	acct, err := s.accountRepo.FindByEmail(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		username, err := s.usernames.UsernameFromEmail(normalized)
		if err != nil {
			return nil, err
		}
		acct, err = s.accountRepo.FindByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if acct == nil {
			return nil, account.ErrUserNotFound
		}
	}

	u, err := s.userRepo.FindByID(ctx, acct.ID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, account.ErrUserNotFound
	}

	return newAdminUserDTO(acct, u, normalized), nil
}

func (s *AdminUserQueryService) ListAdmins(ctx context.Context) ([]AdminUserDTO, error) {
	users, err := s.userRepo.FindByRole(ctx, auth.RoleAdmin)
	if err != nil {
		return nil, err
	}
	items := make([]AdminUserDTO, 0, len(users))
	for i := range users {
		acct, err := s.accountRepo.FindByID(ctx, users[i].ID)
		if err != nil {
			return nil, err
		}
		if acct == nil {
			continue
		}
		items = append(items, *newAdminUserDTO(acct, &users[i], acct.Email))
	}
	return items, nil
}
