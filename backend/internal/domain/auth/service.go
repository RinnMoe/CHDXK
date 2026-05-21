package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"jcourse/internal/domain/task"
)

type RegistrationConfig struct {
	EmailWhitelist []string
	CodeInterval   time.Duration
	CodeTTL        time.Duration
}

type RegistrationService struct {
	userRepo  UserRepository
	codes     VerificationCodeRepository
	sender    VerificationCodeSender
	hasher    PasswordHasher
	whitelist EmailWhitelist
	config    RegistrationConfig
}

func NewRegistrationService(
	userRepo UserRepository,
	codes VerificationCodeRepository,
	sender VerificationCodeSender,
	hasher PasswordHasher,
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
		whitelist: NewEmailWhitelist(config.EmailWhitelist),
		config:    config,
	}
}

func (s *RegistrationService) SendRegisterCode(ctx context.Context, email string) error {
	normalized, err := s.normalizeAllowedEmail(email)
	if err != nil {
		return err
	}
	existing, err := s.userRepo.FindByEmail(ctx, normalized)
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

func (s *RegistrationService) Register(ctx context.Context, email, code, password string) (*User, error) {
	normalized, err := s.normalizeAllowedEmail(email)
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

	existing, err := s.userRepo.FindByEmail(ctx, normalized)
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
	u := NewRegisteredUser(normalized, passwordHash, now)
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}
	_ = s.codes.Delete(ctx, normalized)
	return u, nil
}

func (s *RegistrationService) normalizeAllowedEmail(email string) (string, error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return "", err
	}
	if !s.whitelist.Allows(normalized) {
		return "", ErrEmailNotAllowed
	}
	return normalized, nil
}

type AuthenticationService struct {
	userRepo UserRepository
	hasher   PasswordHasher
}

func NewAuthenticationService(userRepo UserRepository, hasher PasswordHasher) *AuthenticationService {
	return &AuthenticationService{userRepo: userRepo, hasher: hasher}
}

func (s *AuthenticationService) Login(ctx context.Context, email, password string) (*User, error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	u, err := s.userRepo.FindByEmail(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if u == nil || !s.hasher.Verify(password, u.Password) {
		return nil, ErrInvalidCredentials
	}
	if err := EnsureUserActive(ctx, u); err != nil {
		return nil, err
	}
	now := time.Now()
	u.LastSeenAt = now
	if err := s.userRepo.TouchLastSeen(ctx, u.ID, now); err != nil {
		return nil, err
	}
	return u, nil
}

type AuthService struct {
	userRepo UserRepository
}

func NewAuthService(userRepo UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) GetUser(ctx context.Context, id int) (*User, error) {
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}
	if err := EnsureUserActive(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *AuthService) ClearExpiredSuspension(ctx context.Context, userID int) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return nil
	}
	if !u.SuspensionExpired() {
		return nil
	}
	u.ClearSuspension()
	return s.userRepo.Update(ctx, u)
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

func EnsureUserActive(ctx context.Context, u *User) error {
	if u.SuspensionExpired() {
		_ = task.Enqueue(ctx, NewClearExpiredSuspensionTask(u.ID))
		return nil
	}
	if u.IsSuspended() {
		return ErrUserSuspended
	}
	return nil
}
