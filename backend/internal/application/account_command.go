package application

import (
	"context"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
)

type AccountCommandConfig struct {
	Registration  account.RegistrationConfig
	PasswordReset account.PasswordResetConfig
	Login         account.LoginConfig
}

type AccountCommandService struct {
	registration    *account.RegistrationService
	login           *account.LoginService
	passwordReset   *account.PasswordResetService
	authUserService *auth.AuthUserService
}

func NewAccountCommandService(
	accountRepo account.AccountRepository,
	authUserService *auth.AuthUserService,
	codes account.VerificationCodeRepository,
	resetCodes account.VerificationCodeRepository,
	sender account.VerificationCodeSender,
	hasher account.PasswordHasher,
	usernames account.UsernameDeriver,
	config AccountCommandConfig,
	loginAttempts account.LoginAttemptRepository,
) *AccountCommandService {
	return &AccountCommandService{
		registration:    account.NewRegistrationService(accountRepo, codes, sender, hasher, usernames, config.Registration),
		login:           account.NewLoginService(accountRepo, hasher, loginAttempts, usernames, config.Login),
		passwordReset:   account.NewPasswordResetService(accountRepo, resetCodes, sender, hasher, usernames, config.PasswordReset),
		authUserService: authUserService,
	}
}

func (s *AccountCommandService) SendRegisterCode(ctx context.Context, cmd SendRegisterCodeCommand) error {
	return s.registration.SendRegisterCode(ctx, cmd.Email)
}

func (s *AccountCommandService) Register(ctx context.Context, cmd RegisterCommand) (*AccountDTO, error) {
	acct, err := s.registration.Register(ctx, cmd.Email, cmd.Code, cmd.Password)
	if err != nil {
		return nil, err
	}
	u, err := s.authUserService.GetUser(ctx, acct.ID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, account.ErrUserNotFound
	}
	return newAccountDTO(acct, u), nil
}

func (s *AccountCommandService) Login(ctx context.Context, cmd LoginCommand) (*AccountDTO, error) {
	acct, err := s.login.Login(ctx, cmd.Email, cmd.Password)
	if err != nil {
		return nil, err
	}
	u, err := s.authUserService.GetUser(ctx, acct.ID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, account.ErrUserNotFound
	}
	if err := s.login.MarkLogin(ctx, acct.ID); err != nil {
		return nil, err
	}
	return newAccountDTO(acct, u), nil
}

func (s *AccountCommandService) SendResetCode(ctx context.Context, cmd SendResetCodeCommand) error {
	return s.passwordReset.SendResetCode(ctx, cmd.Email)
}

func (s *AccountCommandService) ResetPassword(ctx context.Context, cmd ResetPasswordCommand) error {
	return s.passwordReset.ResetPassword(ctx, cmd.Email, cmd.Code, cmd.NewPassword)
}

func newAccountDTO(acct *account.Account, u *auth.User) *AccountDTO {
	role := ""
	if u != nil {
		role = u.Role
	}
	return &AccountDTO{ID: acct.ID, Username: acct.Username, Email: acct.Email, Role: role}
}
