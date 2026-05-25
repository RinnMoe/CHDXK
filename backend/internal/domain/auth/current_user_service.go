package auth

import (
	"context"
	"time"
)

type AuthUserService struct {
	userRepo UserRepository
}

func NewCurrentUserService(userRepo UserRepository) *AuthUserService {
	return &AuthUserService{userRepo: userRepo}
}

func (s *AuthUserService) GetUser(ctx context.Context, id int) (*User, error) {
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}
	if err := EnsureUserActive(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *AuthUserService) ClearExpiredSuspension(ctx context.Context, userID int) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return nil
	}
	if !u.SuspensionExpired() {
		return nil
	}
	u.ClearSuspension()
	return s.userRepo.Update(ctx, u)
}

func (s *AuthUserService) SuspendUser(ctx context.Context, userID int, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return nil
	}
	if u.IsAdmin() {
		return nil
	}
	u.Suspend(duration)
	return s.userRepo.Update(ctx, u)
}
