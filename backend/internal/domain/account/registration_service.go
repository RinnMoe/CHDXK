package account

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jcourse/internal/domain/account/credential"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/account/notification"
	"jcourse/internal/domain/account/verification"
	"jcourse/internal/domain/email"
)

type RegistrationConfig struct {
	EmailWhitelist []string
}

var DefaultRegistrationConfig = RegistrationConfig{}

type RegistrationService struct {
	accountRepo  identity.Repository
	codes        verification.CodeRepository
	sender       email.Sender
	hasher       credential.PasswordHasher
	usernames    identity.UsernameDeriver
	whitelist    identity.EmailWhitelist
	config       RegistrationConfig
	verification verification.Config
}

func NewRegistrationService(
	accountRepo identity.Repository,
	codes verification.CodeRepository,
	sender email.Sender,
	hasher credential.PasswordHasher,
	usernames identity.UsernameDeriver,
	config RegistrationConfig,
	verificationConfig verification.Config,
) *RegistrationService {
	return &RegistrationService{
		accountRepo:  accountRepo,
		codes:        codes,
		sender:       sender,
		hasher:       hasher,
		usernames:    usernames,
		whitelist:    identity.NewEmailWhitelist(config.EmailWhitelist),
		config:       config,
		verification: verificationConfig.WithDefaults(),
	}
}

func (s *RegistrationService) SendRegisterCode(ctx context.Context, email string) error {
	normalized, err := s.normalizeAllowedEmail(email)
	if err != nil {
		return err
	}
	username, err := s.usernames.UsernameFromEmail(normalized)
	if err != nil {
		return err
	}
	existing, err := s.accountRepo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil {
		return identity.ErrAlreadyExists
	}

	wait, err := s.codes.ReserveSend(ctx, normalized, s.verification.CodeInterval)
	if err != nil {
		return err
	}
	if wait > 0 {
		return fmt.Errorf("%w: retry after %s", verification.ErrSendTooSoon, wait.Round(time.Second))
	}

	code, err := verification.NewCode(normalized, time.Now(), s.verification)
	if err != nil {
		return err
	}
	if err := s.codes.Save(ctx, code, s.verification.CodeTTL); err != nil {
		return err
	}
	mail, err := notification.NewVerificationCodeEmail(normalized, code.Code, s.verification.CodeTTL)
	if err != nil {
		return err
	}
	return s.sender.SendEmail(ctx, mail)
}

func (s *RegistrationService) Register(ctx context.Context, email, code, password string) (*identity.Account, error) {
	normalized, err := s.normalizeAllowedEmail(email)
	if err != nil {
		return nil, err
	}
	username, err := s.usernames.UsernameFromEmail(normalized)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(password) == "" {
		return nil, credential.ErrPasswordRequired
	}

	verificationCode, err := s.codes.Get(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if !verificationCode.Matches(code, time.Now()) {
		return nil, verification.ErrCodeInvalid
	}

	existing, err := s.accountRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, identity.ErrAlreadyExists
	}

	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	acct := identity.NewRegisteredAccount(username, passwordHash, now)
	if err := s.accountRepo.Create(ctx, acct); err != nil {
		return nil, err
	}
	_ = s.codes.Delete(ctx, normalized)
	return acct, nil
}

func (s *RegistrationService) normalizeAllowedEmail(email string) (string, error) {
	normalized := identity.NormalizeEmail(email)
	if !s.whitelist.Allows(normalized) {
		return "", identity.ErrEmailNotAllowed
	}
	return normalized, nil
}
