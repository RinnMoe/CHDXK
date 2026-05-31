package application

import (
	"strings"
	"time"

	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
)

type AdminUserDTO struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Role         string     `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	LastSeenAt   time.Time  `json:"last_seen_at"`
	Suspended    bool       `json:"suspended"`
	SuspendedAt  *time.Time `json:"suspended_at,omitempty"`
	SuspendTill  *time.Time `json:"suspend_till,omitempty"`
	PasswordHash string     `json:"password_hash,omitempty"`
}

func newAdminUserDTO(acct *identity.Account, u *auth.User, lookupEmail string) *AdminUserDTO {
	email := acct.Email
	if email == "" {
		email = lookupEmail
	}
	return &AdminUserDTO{
		ID:           acct.ID,
		Username:     acct.Username,
		Email:        email,
		Role:         u.Role,
		CreatedAt:    acct.CreatedAt,
		LastSeenAt:   acct.LastSeenAt,
		Suspended:    u.IsSuspended(),
		SuspendedAt:  u.SuspendedAt,
		SuspendTill:  u.SuspendTill,
		PasswordHash: maskPasswordHash(acct.PasswordHash),
	}
}

func maskPasswordHash(encoded string) string {
	if encoded == "" {
		return ""
	}

	parts := strings.Split(encoded, "$")
	if len(parts) == 4 && parts[3] != "" {
		parts[2] = maskHashSecret(parts[2])
		parts[3] = maskHashSecret(parts[3])
		return strings.Join(parts, "$")
	}

	return maskHashSecret(encoded)
}

func maskHashSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return value
	}
	return value[:4] + "****"
}
