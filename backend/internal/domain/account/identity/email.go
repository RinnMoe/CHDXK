package identity

import (
	"encoding/hex"
	"strings"

	"jcourse/pkg/apperr"

	"golang.org/x/crypto/blake2b"
)

var ErrEmailNotAllowed = apperr.ErrEmailNotAllowed

type EmailWhitelist struct {
	entries []string
}

type UsernameDeriver interface {
	UsernameFromEmail(email string) (string, error)
}

type UsernameDeriverConfig struct {
	Salt string
}

type BLAKE2bUsernameDeriver struct {
	salt string
}

func NewBLAKE2bUsernameDeriver(config UsernameDeriverConfig) *BLAKE2bUsernameDeriver {
	return &BLAKE2bUsernameDeriver{salt: config.Salt}
}

func NewEmailWhitelist(entries []string) EmailWhitelist {
	normalized := make([]string, 0, len(entries))
	for _, entry := range entries {
		entry = strings.ToLower(strings.TrimSpace(entry))
		if entry != "" {
			normalized = append(normalized, entry)
		}
	}
	return EmailWhitelist{entries: normalized}
}

func (d *BLAKE2bUsernameDeriver) UsernameFromEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	localPart, _, ok := strings.Cut(email, "@")
	if !ok || localPart == "" {
		return "", ErrEmailNotAllowed
	}

	h, err := blake2b.New(16, nil)
	if err != nil {
		return "", err
	}
	_, _ = h.Write([]byte(strings.ToLower(localPart) + d.salt))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (w EmailWhitelist) Allows(email string) bool {
	if len(w.entries) == 0 {
		return false
	}
	email = strings.ToLower(strings.TrimSpace(email))
	for _, entry := range w.entries {
		switch {
		case entry == "*":
			return true
		case strings.HasPrefix(entry, "@") && strings.HasSuffix(email, entry):
			return true
		case email == entry:
			return true
		}
	}
	return false
}
