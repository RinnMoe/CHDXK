package auth

import (
	"net/mail"
	"strings"
)

type EmailWhitelist struct {
	entries []string
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

func NormalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", ErrEmailNotAllowed
	}
	return strings.ToLower(addr.Address), nil
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
