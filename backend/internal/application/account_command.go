package application

import (
	"context"
	"time"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
)

type AccountCommandConfig = account.RegistrationConfig

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
	config AccountCommandConfig,
	resetConfig account.PasswordResetConfig,
	loginAttempts account.LoginAttemptRepository,
	maxLoginAttempts int,
	loginLockout time.Duration,
) *AccountCommandService {
	return &AccountCommandService{
		registration:    account.NewRegistrationService(accountRepo, codes, sender, hasher, config),
		login:           account.NewLoginService(accountRepo, hasher, loginAttempts, maxLoginAttempts, loginLockout),
		passwordReset:   account.NewPasswordResetService(accountRepo, resetCodes, sender, hasher, resetConfig),
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
