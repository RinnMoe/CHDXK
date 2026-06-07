package application

import (
	"context"

	"jcourse/internal/domain/account"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
)

type AccountCommandService struct {
	registration    *account.RegistrationService
	login           *account.LoginService
	passwordReset   *account.PasswordResetService
	authUserService *auth.AuthUserService
	sessionAuth     *auth.SessionAuthService
	settings        SiteSettingsProvider
}

func NewAccountCommandService(
	registration *account.RegistrationService,
	login *account.LoginService,
	passwordReset *account.PasswordResetService,
	authUserService *auth.AuthUserService,
	sessionAuth *auth.SessionAuthService,
	settings SiteSettingsProvider,
) *AccountCommandService {
	if settings == nil {
		defaults := NewDefaultSiteSettingsProvider()
		settings = defaults
	}
	return &AccountCommandService{
		registration:    registration,
		login:           login,
		passwordReset:   passwordReset,
		authUserService: authUserService,
		sessionAuth:     sessionAuth,
		settings:        settings,
	}
}

func (s *AccountCommandService) SendRegisterCode(ctx context.Context, cmd SendRegisterCodeCommand) error {
	config, err := s.settings.AccountRuntimeConfig(ctx)
	if err != nil {
		return err
	}
	return s.registration.SendRegisterCodeWithConfig(ctx, cmd.Email, config.Registration, config.Verification)
}

func (s *AccountCommandService) Register(ctx context.Context, cmd RegisterCommand) (*AccountDTO, error) {
	config, err := s.settings.AccountRuntimeConfig(ctx)
	if err != nil {
		return nil, err
	}
	acct, err := s.registration.RegisterWithConfig(ctx, cmd.Email, cmd.Code, cmd.Password, config.Registration)
	if err != nil {
		return nil, err
	}
	u, err := s.authUserService.GetUser(ctx, acct.ID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, identity.ErrNotFound
	}
	return newAccountDTO(acct, u), nil
}

func (s *AccountCommandService) Login(ctx context.Context, cmd LoginCommand) (*AccountDTO, error) {
	config, err := s.settings.AccountRuntimeConfig(ctx)
	if err != nil {
		return nil, err
	}
	acct, err := s.login.LoginWithConfig(ctx, cmd.Email, cmd.Password, config.Login)
	if err != nil {
		return nil, err
	}
	u, err := s.authUserService.GetUser(ctx, acct.ID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, identity.ErrNotFound
	}
	return newAccountDTO(acct, u), nil
}

func (s *AccountCommandService) SendResetCode(ctx context.Context, cmd SendResetCodeCommand) error {
	config, err := s.settings.AccountRuntimeConfig(ctx)
	if err != nil {
		return err
	}
	return s.passwordReset.SendResetCodeWithConfig(ctx, cmd.Email, config.Verification)
}

func (s *AccountCommandService) ResetPassword(ctx context.Context, cmd ResetPasswordCommand) error {
	return s.passwordReset.ResetPassword(ctx, cmd.Email, cmd.Code, cmd.NewPassword)
}

func (s *AccountCommandService) SessionAuthHash(ctx context.Context, userID int) (string, error) {
	return s.sessionAuth.HashForUser(ctx, userID)
}

func newAccountDTO(acct *identity.Account, u *auth.User) *AccountDTO {
	role := ""
	if u != nil {
		role = u.Role
	}
	return &AccountDTO{ID: acct.ID, Username: acct.Username, Email: acct.Email, Role: role}
}
