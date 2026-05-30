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

type PasswordResetService struct {
	accountRepo  identity.Repository
	codes        verification.CodeRepository
	hasher       credential.PasswordHasher
	usernames    identity.UsernameDeriver
	verification verification.Config
}

func NewPasswordResetService(
	accountRepo identity.Repository,
	codes verification.CodeRepository,
	hasher credential.PasswordHasher,
	usernames identity.UsernameDeriver,
	verificationConfig verification.Config,
) *PasswordResetService {
	return &PasswordResetService{
		accountRepo:  accountRepo,
		codes:        codes,
		hasher:       hasher,
		usernames:    usernames,
		verification: verificationConfig.WithDefaults(),
	}
}

func (s *PasswordResetService) SendResetCode(ctx context.Context, email string) error {
	normalized := identity.NormalizeEmail(email)
	if _, err := s.usernames.UsernameFromEmail(normalized); err != nil {
		return err
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
	return task.Enqueue(ctx, domainemail.NewSendEmailTask(string(notification.EmailTemplateVerificationCode), mail))
}

func (s *PasswordResetService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	normalized := identity.NormalizeEmail(email)
	username, err := s.usernames.UsernameFromEmail(normalized)
	if err != nil {
		return err
	}
	if strings.TrimSpace(newPassword) == "" {
		return credential.ErrPasswordRequired
	}

	verificationCode, err := s.codes.Get(ctx, normalized)
	if err != nil {
		return err
	}
	if !verificationCode.Matches(code, time.Now()) {
		return verification.ErrCodeInvalid
	}

	acct, err := s.accountRepo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if acct == nil {
		return identity.ErrNotFound
	}

	passwordHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	acct.PasswordHash = passwordHash
	if err := s.accountRepo.Update(ctx, acct); err != nil {
		return err
	}
	_ = s.codes.Delete(ctx, normalized)
	return nil
}
