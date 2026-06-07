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
	domainemail "jcourse/internal/domain/email"
	"jcourse/internal/domain/task"
)

type RegistrationConfig struct {
	EmailWhitelist []string
}

var DefaultRegistrationConfig = RegistrationConfig{}

type RegistrationService struct {
	accountRepo  identity.Repository
	codes        verification.CodeRepository
	hasher       credential.PasswordHasher
	usernames    identity.UsernameDeriver
	whitelist    identity.EmailWhitelist
	config       RegistrationConfig
	verification verification.Config
}

func NewRegistrationService(
	accountRepo identity.Repository,
	codes verification.CodeRepository,
	hasher credential.PasswordHasher,
	usernames identity.UsernameDeriver,
	config RegistrationConfig,
	verificationConfig verification.Config,
) *RegistrationService {
	return &RegistrationService{
		accountRepo:  accountRepo,
		codes:        codes,
		hasher:       hasher,
		usernames:    usernames,
		whitelist:    identity.NewEmailWhitelist(config.EmailWhitelist),
		config:       config,
		verification: verificationConfig.WithDefaults(),
	}
}

func (s *RegistrationService) SendRegisterCode(ctx context.Context, email string) error {
	return s.SendRegisterCodeWithConfig(ctx, email, s.config, s.verification)
}

func (s *RegistrationService) SendRegisterCodeWithConfig(ctx context.Context, email string, config RegistrationConfig, verificationConfig verification.Config) error {
	verificationConfig = verificationConfig.WithDefaults()
	normalized, err := s.normalizeAllowedEmail(email, config)
	if err != nil {
		return err
	}

	wait, err := s.codes.ReserveSend(ctx, normalized, verificationConfig.CodeInterval)
	if err != nil {
		return err
	}
	if wait > 0 {
		return fmt.Errorf("%w: retry after %s", verification.ErrSendTooSoon, wait.Round(time.Second))
	}

	code, err := verification.NewCode(normalized, time.Now(), verificationConfig)
	if err != nil {
		return err
	}
	if err := s.codes.Save(ctx, code, verificationConfig.CodeTTL); err != nil {
		return err
	}
	mail, err := notification.NewVerificationCodeEmail(normalized, code.Code, verificationConfig.CodeTTL)
	if err != nil {
		return err
	}
	return task.Enqueue(ctx, domainemail.NewSendEmailTask(string(notification.EmailTemplateVerificationCode), mail))
}

func (s *RegistrationService) Register(ctx context.Context, email, code, password string) (*identity.Account, error) {
	return s.RegisterWithConfig(ctx, email, code, password, s.config)
}

func (s *RegistrationService) RegisterWithConfig(ctx context.Context, email, code, password string, config RegistrationConfig) (*identity.Account, error) {
	normalized, err := s.normalizeAllowedEmail(email, config)
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

func (s *RegistrationService) normalizeAllowedEmail(email string, config RegistrationConfig) (string, error) {
	normalized := identity.NormalizeEmail(email)
	whitelist := s.whitelist
	if len(config.EmailWhitelist) > 0 {
		whitelist = identity.NewEmailWhitelist(config.EmailWhitelist)
	}
	if !whitelist.Allows(normalized) {
		return "", identity.ErrEmailNotAllowed
	}
	return normalized, nil
}
