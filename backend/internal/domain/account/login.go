package account

import (
	"context"
	"fmt"
	"time"
)

type LoginService struct {
	accountRepo AccountRepository
	hasher      PasswordHasher
	attempts    LoginAttemptRepository
	usernames   UsernameDeriver
	maxAttempts int
	lockout     time.Duration
}

func NewLoginService(
	userRepo AccountRepository,
	hasher PasswordHasher,
	attempts LoginAttemptRepository,
	usernames UsernameDeriver,
	maxAttempts int,
	lockout time.Duration,
) *LoginService {
	return &LoginService{
		accountRepo: userRepo,
		hasher:      hasher,
		attempts:    attempts,
		usernames:   usernames,
		maxAttempts: maxAttempts,
		lockout:     lockout,
	}
}

func (s *LoginService) Login(ctx context.Context, email, password string) (*Account, error) {
	normalized := normalizeEmail(email)
	username, err := s.usernames.UsernameFromEmail(normalized)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if s.isLocked(ctx, normalized) {
		return nil, ErrLoginLocked
	}

	u, err := s.accountRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if u == nil || !s.hasher.Verify(password, u.PasswordHash) {
		_ = s.recordFailure(ctx, normalized)
		return nil, ErrInvalidCredentials
	}
	_ = s.attempts.Reset(ctx, normalized)
	return u, nil
}

func (s *LoginService) MarkLogin(ctx context.Context, accountID int) error {
	return s.accountRepo.TouchLastSeen(ctx, accountID, time.Now())
}

func (s *LoginService) isLocked(ctx context.Context, email string) bool {
	if s.maxAttempts <= 0 {
		return false
	}
	count, err := s.attempts.Get(ctx, email)
	if err != nil {
		return false
	}
	return count >= s.maxAttempts
}

func (s *LoginService) recordFailure(ctx context.Context, email string) error {
	count, err := s.attempts.Increment(ctx, email)
	if err != nil {
		return err
	}
	if count >= s.maxAttempts && s.lockout > 0 {
		return fmt.Errorf("%w: locked for %s", ErrLoginLocked, s.lockout.Round(time.Minute))
	}
	return nil
}
