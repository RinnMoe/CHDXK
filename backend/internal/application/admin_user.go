package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
)

var (
	ErrCannotSuspendAdmin = errors.New("cannot suspend admin user")
	ErrCannotOperateSelf  = errors.New("cannot operate on yourself")
)

type AdminUserQueryService struct {
	accountRepo account.AccountRepository
	userRepo    auth.UserRepository
	usernames   account.UsernameDeriver
}

type AdminUserCommandService struct {
	userRepo auth.UserRepository
	config   AdminUserCommandConfig
}

type AdminUserCommandConfig struct {
	DefaultSuspendDays int
}

var DefaultAdminUserCommandConfig = AdminUserCommandConfig{DefaultSuspendDays: auth.DefaultAdminConfig.DefaultSuspendDays}

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

func NewAdminUserCommandService(userRepo auth.UserRepository, config AdminUserCommandConfig) *AdminUserCommandService {
	if config.DefaultSuspendDays <= 0 {
		config.DefaultSuspendDays = DefaultAdminUserCommandConfig.DefaultSuspendDays
	}
	return &AdminUserCommandService{userRepo: userRepo, config: config}
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

func (s *AdminUserCommandService) SuspendUserForDays(ctx context.Context, actorUserID int, userID int, days int) error {
	if actorUserID == userID {
		return ErrCannotOperateSelf
	}

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
		days = s.config.DefaultSuspendDays
	}

	u.Suspend(time.Duration(days) * 24 * time.Hour)
	return s.userRepo.Update(ctx, u)
}

func (s *AdminUserCommandService) ClearSuspension(ctx context.Context, actorUserID int, userID int) error {
	if actorUserID == userID {
		return ErrCannotOperateSelf
	}

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

func (s *AdminUserCommandService) GrantAdmin(ctx context.Context, actorUserID int, userID int) error {
	if actorUserID == userID {
		return ErrCannotOperateSelf
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return account.ErrUserNotFound
	}

	u.Role = auth.RoleAdmin
	u.ClearSuspension()
	return s.userRepo.Update(ctx, u)
}

func (s *AdminUserCommandService) RevokeAdmin(ctx context.Context, actorUserID int, userID int) error {
	if actorUserID == userID {
		return ErrCannotOperateSelf
	}

	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return account.ErrUserNotFound
	}

	u.Role = auth.RoleUser
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
