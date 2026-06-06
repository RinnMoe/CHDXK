package account

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/account/security"
	"jcourse/pkg/apperr"
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
	if acct == nil {
		return nil, s.recordFailure(ctx, normalized)
	}
	if strings.TrimSpace(acct.PasswordHash) == "" {
		return nil, security.ErrPasswordNotSet
	}
	if !s.hasher.Verify(password, acct.PasswordHash) {
		return nil, s.recordFailure(ctx, normalized)
	}
	_ = s.attempts.Reset(ctx, normalized)
	return acct, nil
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
		return security.ErrInvalidCredentials
	}
	remaining := max(s.config.MaxAttempts-count, 0)
	msg := fmt.Sprintf("邮箱或密码错误，还有 %d 次尝试机会", remaining)
	return fmt.Errorf("%w: %w", apperr.Unauthorized(msg), security.ErrInvalidCredentials)
}
