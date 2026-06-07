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
}

func NewLoginService(
	accountRepo identity.Repository,
	hasher credential.PasswordHasher,
	attempts security.LoginAttemptRepository,
	usernames identity.UsernameDeriver,
) *LoginService {
	return &LoginService{
		accountRepo: accountRepo,
		hasher:      hasher,
		attempts:    attempts,
		usernames:   usernames,
	}
}

func (s *LoginService) LoginWithConfig(ctx context.Context, email, password string, config LoginConfig) (*identity.Account, error) {
	normalized := identity.NormalizeEmail(email)
	username, err := s.usernames.UsernameFromEmail(normalized)
	if err != nil {
		return nil, security.ErrInvalidCredentials
	}

	if s.isLocked(ctx, normalized, config) {
		return nil, security.ErrLoginLocked
	}

	acct, err := s.accountRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, s.recordFailure(ctx, normalized, config)
	}
	if strings.TrimSpace(acct.PasswordHash) == "" {
		return nil, security.ErrPasswordNotSet
	}
	if !s.hasher.Verify(password, acct.PasswordHash) {
		return nil, s.recordFailure(ctx, normalized, config)
	}
	_ = s.attempts.Reset(ctx, normalized)
	return acct, nil
}

func (s *LoginService) isLocked(ctx context.Context, email string, config LoginConfig) bool {
	if config.MaxAttempts <= 0 {
		return false
	}
	count, err := s.attempts.Get(ctx, email)
	if err != nil {
		return false
	}
	return count >= config.MaxAttempts
}

func (s *LoginService) recordFailure(ctx context.Context, email string, config LoginConfig) error {
	count, err := s.attempts.Increment(ctx, email, config.Lockout)
	if err != nil {
		return security.ErrInvalidCredentials
	}
	remaining := max(config.MaxAttempts-count, 0)
	msg := fmt.Sprintf("邮箱或密码错误，还有 %d 次尝试机会", remaining)
	return fmt.Errorf("%w: %w", apperr.Unauthorized(msg), security.ErrInvalidCredentials)
}
