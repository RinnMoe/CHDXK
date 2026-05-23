package auth

import (
	"context"
)

type CurrentUserService struct {
	userRepo UserRepository
}

func NewCurrentUserService(userRepo UserRepository) *CurrentUserService {
	return &CurrentUserService{userRepo: userRepo}
}

func (s *CurrentUserService) GetUser(ctx context.Context, id int) (*User, error) {
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

func (s *CurrentUserService) ClearExpiredSuspension(ctx context.Context, userID int) error {
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
