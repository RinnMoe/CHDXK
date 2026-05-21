package application

import (
	"context"
	"time"

	"jcourse/internal/domain/auth"
)

type AuthCommandConfig = auth.RegistrationConfig

type AuthCommandService struct {
	registration   *auth.RegistrationService
	authentication *auth.AuthenticationService
	passwordReset  *auth.PasswordResetService
}

func NewAuthCommandService(
	userRepo auth.UserRepository,
	codes auth.VerificationCodeRepository,
	resetCodes auth.VerificationCodeRepository,
	sender auth.VerificationCodeSender,
	hasher auth.PasswordHasher,
	config AuthCommandConfig,
	resetConfig auth.PasswordResetConfig,
	loginAttempts auth.LoginAttemptRepository,
	maxLoginAttempts int,
	loginLockout time.Duration,
) *AuthCommandService {
	return &AuthCommandService{
		registration:   auth.NewRegistrationService(userRepo, codes, sender, hasher, config),
		authentication: auth.NewAuthenticationService(userRepo, hasher, loginAttempts, maxLoginAttempts, loginLockout),
		passwordReset:  auth.NewPasswordResetService(userRepo, resetCodes, sender, hasher, resetConfig),
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

func (s *AuthCommandService) SendResetCode(ctx context.Context, cmd SendResetCodeCommand) error {
	return s.passwordReset.SendResetCode(ctx, cmd.Email)
}

func (s *AuthCommandService) ResetPassword(ctx context.Context, cmd ResetPasswordCommand) error {
	return s.passwordReset.ResetPassword(ctx, cmd.Email, cmd.Code, cmd.NewPassword)
}

func newAuthUserDTO(u *auth.User) *AuthUserDTO {
	return &AuthUserDTO{ID: u.ID, Username: u.Username, Email: u.Email, Role: u.Role}
}
