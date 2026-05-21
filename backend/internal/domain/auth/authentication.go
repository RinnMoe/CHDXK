package auth

import (
	"context"
	"fmt"
	"time"
)

type AuthenticationService struct {
	userRepo    UserRepository
	hasher      PasswordHasher
	attempts    LoginAttemptRepository
	maxAttempts int
	lockout     time.Duration
}

func NewAuthenticationService(
	userRepo UserRepository,
	hasher PasswordHasher,
	attempts LoginAttemptRepository,
	maxAttempts int,
	lockout time.Duration,
) *AuthenticationService {
	return &AuthenticationService{
		userRepo:    userRepo,
		hasher:      hasher,
		attempts:    attempts,
		maxAttempts: maxAttempts,
		lockout:     lockout,
	}
}

func (s *AuthenticationService) Login(ctx context.Context, email, password string) (*User, error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if s.isLocked(ctx, normalized) {
		return nil, ErrLoginLocked
	}

	u, err := s.userRepo.FindByEmail(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if u == nil || !s.hasher.Verify(password, u.Password) {
		_ = s.recordFailure(ctx, normalized)
		return nil, ErrInvalidCredentials
	}
	if err := EnsureUserActive(ctx, u); err != nil {
		return nil, err
	}
	_ = s.attempts.Reset(ctx, normalized)

	now := time.Now()
	u.LastSeenAt = now
	if err := s.userRepo.TouchLastSeen(ctx, u.ID, now); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *AuthenticationService) isLocked(ctx context.Context, email string) bool {
	if s.maxAttempts <= 0 {
		return false
	}
	count, err := s.attempts.Get(ctx, email)
	if err != nil {
		return false
	}
	return count >= s.maxAttempts
}

func (s *AuthenticationService) recordFailure(ctx context.Context, email string) error {
	count, err := s.attempts.Increment(ctx, email)
	if err != nil {
		return err
	}
	if count >= s.maxAttempts && s.lockout > 0 {
		return fmt.Errorf("%w: locked for %s", ErrLoginLocked, s.lockout.Round(time.Minute))
	}
	return nil
}
