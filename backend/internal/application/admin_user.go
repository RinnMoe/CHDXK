package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
)

var ErrCannotSuspendAdmin = errors.New("cannot suspend admin user")

type AdminUserQueryService struct {
	accountRepo account.AccountRepository
	userRepo    auth.UserRepository
	usernames   account.UsernameDeriver
}

type AdminUserCommandService struct {
	userRepo auth.UserRepository
}

type AdminUserDTO struct {
	ID          int        `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	CreatedAt   time.Time  `json:"created_at"`
	LastSeenAt  time.Time  `json:"last_seen_at"`
	Suspended   bool       `json:"suspended"`
	SuspendedAt *time.Time `json:"suspended_at,omitempty"`
	SuspendTill *time.Time `json:"suspend_till,omitempty"`
}

func NewAdminUserQueryService(
	accountRepo account.AccountRepository,
	userRepo auth.UserRepository,
	usernames account.UsernameDeriver,
) *AdminUserQueryService {
	return &AdminUserQueryService{accountRepo: accountRepo, userRepo: userRepo, usernames: usernames}
}

func NewAdminUserCommandService(userRepo auth.UserRepository) *AdminUserCommandService {
	return &AdminUserCommandService{userRepo: userRepo}
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

func (s *AdminUserCommandService) SuspendUser(ctx context.Context, userID int) error {
	return s.SuspendUserForDays(ctx, userID, 30)
}

func (s *AdminUserCommandService) SuspendUserForDays(ctx context.Context, userID int, days int) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return account.ErrUserNotFound
	}
	if u.IsAdmin() {
		return ErrCannotSuspendAdmin
	}
	if days <= 0 {
		days = 30
	}

	u.Suspend(time.Duration(days) * 24 * time.Hour)
	return s.userRepo.Update(ctx, u)
}

func (s *AdminUserCommandService) ClearSuspension(ctx context.Context, userID int) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return account.ErrUserNotFound
	}

	u.ClearSuspension()
	return s.userRepo.Update(ctx, u)
}

func newAdminUserDTO(acct *account.Account, u *auth.User, lookupEmail string) *AdminUserDTO {
	email := acct.Email
	if email == "" {
		email = lookupEmail
	}
	return &AdminUserDTO{
		ID:          acct.ID,
		Username:    acct.Username,
		Email:       email,
		Role:        u.Role,
		CreatedAt:   acct.CreatedAt,
		LastSeenAt:  acct.LastSeenAt,
		Suspended:   u.IsSuspended(),
		SuspendedAt: u.SuspendedAt,
		SuspendTill: u.SuspendTill,
	}
}
