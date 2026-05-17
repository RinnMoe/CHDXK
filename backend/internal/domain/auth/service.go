package auth

import "context"

type AuthService struct {
	userRepo UserRepository
}

func NewAuthService(userRepo UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) GetUser(ctx context.Context, id int) (*User, error) {
	return s.userRepo.FindByID(ctx, id)
}
