package account

import (
	"context"
	"fmt"
	"time"

	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/account/security"
)

type LoginConfig struct {
	MaxAttempts int
	Lockout     time.Duration
}

var DefaultLoginConfig = LoginConfig{
	MaxAttempts: 5,
	Lockout:     15 * time.Minute,
}

type LoginService struct {
	accountRepo identity.Repository
	hasher      credential.PasswordHasher
	attempts    security.LoginAttemptRepository
	usernames   identity.UsernameDeriver
	config      LoginConfig
}

func NewLoginService(
	accountRepo identity.Repository,
	hasher credential.PasswordHasher,
	attempts security.LoginAttemptRepository,
	usernames identity.UsernameDeriver,
	config LoginConfig,
) *LoginService {
	defaults := DefaultLoginConfig
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = defaults.MaxAttempts
	}
	if config.Lockout <= 0 {
		config.Lockout = defaults.Lockout
	}
	return &LoginService{
		accountRepo: accountRepo,
		hasher:      hasher,
		attempts:    attempts,
		usernames:   usernames,
		config:      config,
	}
}

func (s *LoginService) Login(ctx context.Context, email, password string) (*identity.Account, error) {
	normalized := identity.NormalizeEmail(email)
	username, err := s.usernames.UsernameFromEmail(normalized)
	if err != nil {
		return nil, security.ErrInvalidCredentials
	}

	if s.isLocked(ctx, normalized) {
		return nil, security.ErrLoginLocked
	}

	acct, err := s.accountRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if acct == nil || !s.hasher.Verify(password, acct.PasswordHash) {
		_ = s.recordFailure(ctx, normalized)
		return nil, security.ErrInvalidCredentials
	}
	_ = s.attempts.Reset(ctx, normalized)
	return acct, nil
}

func (s *LoginService) MarkLogin(ctx context.Context, accountID int) error {
	return s.accountRepo.TouchLastSeen(ctx, accountID, time.Now())
}

func (s *LoginService) isLocked(ctx context.Context, email string) bool {
	if s.config.MaxAttempts <= 0 {
		return false
	}
	count, err := s.attempts.Get(ctx, email)
	if err != nil {
		return false
	}
	return count >= s.config.MaxAttempts
}

func (s *LoginService) recordFailure(ctx context.Context, email string) error {
	count, err := s.attempts.Increment(ctx, email)
	if err != nil {
		return err
	}
	if count >= s.config.MaxAttempts && s.config.Lockout > 0 {
		return fmt.Errorf("%w: locked for %s", security.ErrLoginLocked, s.config.Lockout.Round(time.Minute))
	}
	return nil
}
