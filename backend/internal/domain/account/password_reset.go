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
}

type PasswordResetService struct {
	userRepo AccountRepository
	codes    VerificationCodeRepository
	sender   VerificationCodeSender
	hasher   PasswordHasher
	config   PasswordResetConfig
}

func NewPasswordResetService(
	userRepo AccountRepository,
	codes VerificationCodeRepository,
	sender VerificationCodeSender,
	hasher PasswordHasher,
	config PasswordResetConfig,
) *PasswordResetService {
	if config.CodeInterval <= 0 {
		config.CodeInterval = time.Minute
	}
	if config.CodeTTL <= 0 {
		config.CodeTTL = 10 * time.Minute
	}
	return &PasswordResetService{
		userRepo: userRepo,
		codes:    codes,
		sender:   sender,
		hasher:   hasher,
		config:   config,
	}
}

func (s *PasswordResetService) SendResetCode(ctx context.Context, email string) error {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return err
	}

	existing, err := s.userRepo.FindByEmail(ctx, normalized)
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

func (s *PasswordResetService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	normalized, err := NormalizeEmail(email)
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

	u, err := s.userRepo.FindByEmail(ctx, normalized)
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
	u.Password = passwordHash
	if err := s.userRepo.Update(ctx, u); err != nil {
		return err
	}
	_ = s.codes.Delete(ctx, normalized)
	return nil
}
