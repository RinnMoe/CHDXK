package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"jcourse/internal/domain/account/identity"
	"jcourse/pkg/apperr"
)

var ErrInvalidSession = apperr.Unauthorized("登录状态已失效，请重新登录")

type SessionAuthService struct {
	accountRepo identity.Repository
	secret      []byte
}

func NewSessionAuthService(accountRepo identity.Repository, secret string) *SessionAuthService {
	return &SessionAuthService{accountRepo: accountRepo, secret: []byte(secret)}
}

func (s *SessionAuthService) HashForUser(ctx context.Context, userID int) (string, error) {
	if s == nil || s.accountRepo == nil || len(s.secret) == 0 {
		return "", ErrInvalidSession
	}
	acct, err := s.accountRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if acct == nil || strings.TrimSpace(acct.PasswordHash) == "" {
		return "", ErrInvalidSession
	}
	return s.hash(userID, acct.PasswordHash), nil
}

func (s *SessionAuthService) Validate(ctx context.Context, userID int, sessionHash string) error {
	if strings.TrimSpace(sessionHash) == "" {
		return ErrInvalidSession
	}
	expected, err := s.HashForUser(ctx, userID)
	if err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(expected), []byte(sessionHash)) != 1 {
		return ErrInvalidSession
	}
	return nil
}

func (s *SessionAuthService) hash(userID int, passwordHash string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(strconv.Itoa(userID)))
	_, _ = mac.Write([]byte(":"))
	_, _ = mac.Write([]byte(passwordHash))
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(mac.Sum(nil)))
}
