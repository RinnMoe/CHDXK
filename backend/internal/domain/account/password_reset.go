package account

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type PasswordResetConfig struct {
	CodeInterval time.Duration
	CodeTTL      time.Duration
	CodeLength   int
}

func DefaultPasswordResetConfig() PasswordResetConfig {
	return PasswordResetConfig{
		CodeInterval: time.Minute,
		CodeTTL:      10 * time.Minute,
		CodeLength:   6,
	}
}

type PasswordResetService struct {
	userRepo  AccountRepository
	codes     VerificationCodeRepository
	sender    VerificationCodeSender
	hasher    PasswordHasher
	usernames UsernameDeriver
	config    PasswordResetConfig
}

func NewPasswordResetService(
	userRepo AccountRepository,
	codes VerificationCodeRepository,
	sender VerificationCodeSender,
	hasher PasswordHasher,
	usernames UsernameDeriver,
	config PasswordResetConfig,
) *PasswordResetService {
	defaults := DefaultPasswordResetConfig()
	if config.CodeInterval <= 0 {
		config.CodeInterval = defaults.CodeInterval
	}
	if config.CodeTTL <= 0 {
		config.CodeTTL = defaults.CodeTTL
	}
	if config.CodeLength <= 0 {
		config.CodeLength = defaults.CodeLength
	}
	return &PasswordResetService{
		userRepo:  userRepo,
		codes:     codes,
		sender:    sender,
		hasher:    hasher,
		usernames: usernames,
		config:    config,
	}
}

func (s *PasswordResetService) SendResetCode(ctx context.Context, email string) error {
	normalized := normalizeEmail(email)
	username, err := s.usernames.UsernameFromEmail(normalized)
	if err != nil {
		return err
	}

	existing, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrUserNotFound
	}

	wait, err := s.codes.ReserveSend(ctx, normalized, s.config.CodeInterval)
	if err != nil {
		return err
	}
	if wait > 0 {
		return fmt.Errorf("%w: retry after %s", ErrVerificationTooSoon, wait.Round(time.Second))
	}

	code, err := numericCode(s.config.CodeLength)
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

func (s *PasswordResetService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	normalized := normalizeEmail(email)
	username, err := s.usernames.UsernameFromEmail(normalized)
	if err != nil {
		return err
	}
	if strings.TrimSpace(newPassword) == "" {
		return ErrPasswordRequired
	}

	verificationCode, err := s.codes.Get(ctx, normalized)
	if err != nil {
		return err
	}
	if !verificationCode.Matches(code, time.Now()) {
		return ErrVerificationCodeInvalid
	}

	u, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrUserNotFound
	}

	passwordHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	u.PasswordHash = passwordHash
	if err := s.userRepo.Update(ctx, u); err != nil {
		return err
	}
	_ = s.codes.Delete(ctx, normalized)
	return nil
}
