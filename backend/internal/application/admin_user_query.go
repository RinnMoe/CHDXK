package application

import (
	"context"
	"strings"

	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/review"
)

type AdminUserQueryService struct {
	accountRepo identity.Repository
	userRepo    auth.UserRepository
	reviewQuery review.ReviewQuery
	usernames   identity.UsernameDeriver
}

type AdminUserLookup struct {
	Email    string
	Username string
	ReviewID int
}

func NewAdminUserQueryService(
	accountRepo identity.Repository,
	userRepo auth.UserRepository,
	reviewQuery review.ReviewQuery,
	usernames identity.UsernameDeriver,
) *AdminUserQueryService {
	return &AdminUserQueryService{accountRepo: accountRepo, userRepo: userRepo, reviewQuery: reviewQuery, usernames: usernames}
}

func (s *AdminUserQueryService) FindUser(ctx context.Context, lookup AdminUserLookup) (*AdminUserDTO, error) {
	if strings.TrimSpace(lookup.Email) != "" {
		return s.findByEmail(ctx, lookup.Email)
	}
	if strings.TrimSpace(lookup.Username) != "" {
		return s.findByUsername(ctx, lookup.Username)
	}
	if lookup.ReviewID > 0 {
		return s.findByReviewID(ctx, lookup.ReviewID)
	}
	return nil, identity.ErrNotFound
}

func (s *AdminUserQueryService) findByEmail(ctx context.Context, email string) (*AdminUserDTO, error) {
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
			return nil, identity.ErrNotFound
		}
	}

	return s.adminUserDTOByAccount(ctx, acct, normalized)
}

func (s *AdminUserQueryService) findByUsername(ctx context.Context, username string) (*AdminUserDTO, error) {
	normalized := strings.TrimSpace(username)
	acct, err := s.accountRepo.FindByUsername(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, identity.ErrNotFound
	}

	return s.adminUserDTOByAccount(ctx, acct, acct.Email)
}

func (s *AdminUserQueryService) findByReviewID(ctx context.Context, reviewID int) (*AdminUserDTO, error) {
	rv, err := s.reviewQuery.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if rv == nil {
		return nil, identity.ErrNotFound
	}

	acct, err := s.accountRepo.FindByID(ctx, rv.UserID)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, identity.ErrNotFound
	}

	return s.adminUserDTOByAccount(ctx, acct, acct.Email)
}

func (s *AdminUserQueryService) adminUserDTOByAccount(ctx context.Context, acct *identity.Account, lookupEmail string) (*AdminUserDTO, error) {
	u, err := s.userRepo.FindByID(ctx, acct.ID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, identity.ErrNotFound
	}

	return newAdminUserDTO(acct, u, lookupEmail), nil
}

func (s *AdminUserQueryService) ListAdmins(ctx context.Context) ([]AdminUserDTO, error) {
	users, err := s.userRepo.FindAdmin(ctx)
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
