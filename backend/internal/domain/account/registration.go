package account

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type RegistrationConfig struct {
	EmailWhitelist []string
	CodeInterval   time.Duration
	CodeTTL        time.Duration
}

type RegistrationService struct {
	userRepo  AccountRepository
	codes     VerificationCodeRepository
	sender    VerificationCodeSender
	hasher    PasswordHasher
	usernames UsernameDeriver
	whitelist EmailWhitelist
	config    RegistrationConfig
}

func NewRegistrationService(
	userRepo AccountRepository,
	codes VerificationCodeRepository,
	sender VerificationCodeSender,
	hasher PasswordHasher,
	usernames UsernameDeriver,
	config RegistrationConfig,
) *RegistrationService {
	if config.CodeInterval <= 0 {
		config.CodeInterval = time.Minute
	}
	if config.CodeTTL <= 0 {
		config.CodeTTL = 10 * time.Minute
	}
	return &RegistrationService{
		userRepo:  userRepo,
		codes:     codes,
		sender:    sender,
		hasher:    hasher,
		usernames: usernames,
		whitelist: NewEmailWhitelist(config.EmailWhitelist),
		config:    config,
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
	existing, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrUserAlreadyExists
	}

	wait, err := s.codes.ReserveSend(ctx, normalized, s.config.CodeInterval)
	if err != nil {
		return err
	}
	if wait > 0 {
		return fmt.Errorf("%w: retry after %s", ErrVerificationTooSoon, wait.Round(time.Second))
	}

	code, err := numericCode(6)
	if err != nil {
		return err
	}
	if err := s.codes.Save(ctx, VerificationCode{
		Email:     normalized,
		Code:      code,
		ExpiresAt: time.Now().Add(s.config.CodeTTL),
	}, s.config.CodeTTL); err != nil {
		return err
	}
	return s.sender.SendVerificationCode(ctx, normalized, code)
}

func (s *RegistrationService) Register(ctx context.Context, email, code, password string) (*Account, error) {
	normalized, err := s.normalizeAllowedEmail(email)
	if err != nil {
		return nil, err
	}
	username, err := s.usernames.UsernameFromEmail(normalized)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(password) == "" {
		return nil, ErrPasswordRequired
	}

	verificationCode, err := s.codes.Get(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if !verificationCode.Matches(code, time.Now()) {
		return nil, ErrVerificationCodeInvalid
	}

	existing, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	u := NewRegisteredAccount(username, passwordHash, now)
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}
	_ = s.codes.Delete(ctx, normalized)
	return u, nil
}

func (s *RegistrationService) normalizeAllowedEmail(email string) (string, error) {
	normalized := normalizeEmail(email)
	if !s.whitelist.Allows(normalized) {
		return "", ErrEmailNotAllowed
	}
	return normalized, nil
}
