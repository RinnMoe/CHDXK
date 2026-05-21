package auth

import (
	"context"
	"time"
)

type AuthenticationService struct {
	userRepo UserRepository
	hasher   PasswordHasher
}

func NewAuthenticationService(userRepo UserRepository, hasher PasswordHasher) *AuthenticationService {
	return &AuthenticationService{userRepo: userRepo, hasher: hasher}
}

func (s *AuthenticationService) Login(ctx context.Context, email, password string) (*User, error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	u, err := s.userRepo.FindByEmail(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if u == nil || !s.hasher.Verify(password, u.Password) {
		return nil, ErrInvalidCredentials
	}
	if err := EnsureUserActive(ctx, u); err != nil {
		return nil, err
	}
	now := time.Now()
	u.LastSeenAt = now
	if err := s.userRepo.TouchLastSeen(ctx, u.ID, now); err != nil {
		return nil, err
	}
	return u, nil
}
