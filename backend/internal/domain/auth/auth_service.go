package auth

import (
	"context"
)

type AuthService struct {
	userRepo UserRepository
}

func NewAuthService(userRepo UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) GetUser(ctx context.Context, id int) (*User, error) {
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

func (s *AuthService) ClearExpiredSuspension(ctx context.Context, userID int) error {
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
