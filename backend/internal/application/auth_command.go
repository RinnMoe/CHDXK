package application

import (
	"context"

	"jcourse/internal/domain/auth"
)

type AuthCommandConfig = auth.RegistrationConfig

type AuthCommandService struct {
	registration   *auth.RegistrationService
	authentication *auth.AuthenticationService
}

func NewAuthCommandService(
	userRepo auth.UserRepository,
	codes auth.VerificationCodeRepository,
	sender auth.VerificationCodeSender,
	hasher auth.PasswordHasher,
	config AuthCommandConfig,
) *AuthCommandService {
	return &AuthCommandService{
		registration:   auth.NewRegistrationService(userRepo, codes, sender, hasher, config),
		authentication: auth.NewAuthenticationService(userRepo, hasher),
	}
}

func (s *AuthCommandService) SendRegisterCode(ctx context.Context, cmd SendRegisterCodeCommand) error {
	return s.registration.SendRegisterCode(ctx, cmd.Email)
}

func (s *AuthCommandService) Register(ctx context.Context, cmd RegisterCommand) (*AuthUserDTO, error) {
	u, err := s.registration.Register(ctx, cmd.Email, cmd.Code, cmd.Password)
	if err != nil {
		return nil, err
	}
	return newAuthUserDTO(u), nil
}

func (s *AuthCommandService) Login(ctx context.Context, cmd LoginCommand) (*AuthUserDTO, error) {
	u, err := s.authentication.Login(ctx, cmd.Email, cmd.Password)
	if err != nil {
		return nil, err
	}
	return newAuthUserDTO(u), nil
}

func newAuthUserDTO(u *auth.User) *AuthUserDTO {
	return &AuthUserDTO{ID: u.ID, Username: u.Username, Email: u.Email, Role: u.Role}
}
