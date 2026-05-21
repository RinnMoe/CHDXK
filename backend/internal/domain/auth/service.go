package auth

import (
	"context"

	"jcourse/internal/domain/task"
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
	if !u.SuspensionExpired() {
		return nil
	}
	u.ClearSuspension()
	return s.userRepo.Update(ctx, u)
}

func EnsureUserActive(ctx context.Context, u *User) error {
	if u.SuspensionExpired() {
		_ = task.Enqueue(ctx, NewClearExpiredSuspensionTask(u.ID))
		return nil
	}
	if u.IsSuspended() {
		return ErrUserSuspended
	}
	return nil
}
