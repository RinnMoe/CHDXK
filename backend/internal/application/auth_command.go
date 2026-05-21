package application

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/auth"
)

type AuthCommandConfig struct {
	EmailWhitelist []string
	CodeInterval   time.Duration
	CodeTTL        time.Duration
}

type AuthCommandService struct {
	userRepo  auth.UserRepository
	codes     auth.VerificationCodeRepository
	sender    auth.VerificationCodeSender
	hasher    auth.PasswordHasher
	whitelist auth.EmailWhitelist
	config    AuthCommandConfig
}

func NewAuthCommandService(
	userRepo auth.UserRepository,
	codes auth.VerificationCodeRepository,
	sender auth.VerificationCodeSender,
	hasher auth.PasswordHasher,
	config AuthCommandConfig,
) *AuthCommandService {
	if config.CodeInterval <= 0 {
		config.CodeInterval = time.Minute
	}
	if config.CodeTTL <= 0 {
		config.CodeTTL = 10 * time.Minute
	}
	return &AuthCommandService{
		userRepo:  userRepo,
		codes:     codes,
		sender:    sender,
		hasher:    hasher,
		whitelist: auth.NewEmailWhitelist(config.EmailWhitelist),
		config:    config,
	}
}

func (s *AuthCommandService) SendRegisterCode(ctx context.Context, cmd SendRegisterCodeCommand) error {
	email, err := s.normalizeAllowedEmail(cmd.Email)
	if err != nil {
		return err
	}
	if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		return auth.ErrUserAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if wait, err := s.codes.ReserveSend(ctx, email, s.config.CodeInterval); err != nil {
		return err
	} else if wait > 0 {
		return fmt.Errorf("%w: retry after %s", auth.ErrVerificationTooSoon, wait.Round(time.Second))
	}

	code, err := numericCode(6)
	if err != nil {
		return err
	}
	if err := s.codes.Save(ctx, auth.VerificationCode{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(s.config.CodeTTL),
	}, s.config.CodeTTL); err != nil {
		return err
	}
	return s.sender.SendVerificationCode(ctx, email, code)
}

func (s *AuthCommandService) Register(ctx context.Context, cmd RegisterCommand) (*AuthUserDTO, error) {
	email, err := s.normalizeAllowedEmail(cmd.Email)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cmd.Password) == "" {
		return nil, errors.New("password is required")
	}

	code, err := s.codes.Get(ctx, email)
	if err != nil {
		return nil, err
	}
	if code == nil || time.Now().After(code.ExpiresAt) || subtle.ConstantTimeCompare([]byte(code.Code), []byte(strings.TrimSpace(cmd.Code))) != 1 {
		return nil, auth.ErrVerificationCodeInvalid
	}
	if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		return nil, auth.ErrUserAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	passwordHash, err := s.hasher.Hash(cmd.Password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	u := &auth.User{
		Username:   email,
		Email:      email,
		Role:       auth.RoleUser,
		Password:   passwordHash,
		CreatedAt:  now,
		LastSeenAt: now,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}
	_ = s.codes.Delete(ctx, email)
	return newAuthUserDTO(u), nil
}

func (s *AuthCommandService) Login(ctx context.Context, cmd LoginCommand) (*AuthUserDTO, error) {
	email, err := auth.NormalizeEmail(cmd.Email)
	if err != nil {
		return nil, auth.ErrInvalidCredentials
	}
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, auth.ErrInvalidCredentials
		}
		return nil, err
	}
	if !s.hasher.Verify(cmd.Password, u.Password) {
		return nil, auth.ErrInvalidCredentials
	}
	if err := auth.EnsureUserActive(ctx, u); err != nil {
		return nil, err
	}
	now := time.Now()
	u.LastSeenAt = now
	if err := s.userRepo.TouchLastSeen(ctx, u.ID, now); err != nil {
		return nil, err
	}
	return newAuthUserDTO(u), nil
}

func (s *AuthCommandService) normalizeAllowedEmail(email string) (string, error) {
	normalized, err := auth.NormalizeEmail(email)
	if err != nil {
		return "", err
	}
	if !s.whitelist.Allows(normalized) {
		return "", auth.ErrEmailNotAllowed
	}
	return normalized, nil
}

func newAuthUserDTO(u *auth.User) *AuthUserDTO {
	return &AuthUserDTO{ID: u.ID, Username: u.Username, Email: u.Email, Role: u.Role}
}

func numericCode(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := make([]byte, length)
	for i, b := range buf {
		code[i] = byte('0' + int(b)%10)
	}
	return string(code), nil
}
