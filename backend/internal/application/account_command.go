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
}

func NewAccountCommandService(
	registration *account.RegistrationService,
	login *account.LoginService,
	passwordReset *account.PasswordResetService,
	authUserService *auth.AuthUserService,
) *AccountCommandService {
	return &AccountCommandService{
		registration:    registration,
		login:           login,
		passwordReset:   passwordReset,
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
		return nil, identity.ErrNotFound
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
		return nil, identity.ErrNotFound
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

func newAccountDTO(acct *identity.Account, u *auth.User) *AccountDTO {
	role := ""
	if u != nil {
		role = u.Role
	}
	return &AccountDTO{ID: acct.ID, Username: acct.Username, Email: acct.Email, Role: role}
}
