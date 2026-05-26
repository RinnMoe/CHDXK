package application

import (
	"time"

	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
)

type AdminUserDTO struct {
	ID          int        `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	CreatedAt   time.Time  `json:"created_at"`
	LastSeenAt  time.Time  `json:"last_seen_at"`
	Suspended   bool       `json:"suspended"`
	SuspendedAt *time.Time `json:"suspended_at,omitempty"`
	SuspendTill *time.Time `json:"suspend_till,omitempty"`
}

func newAdminUserDTO(acct *identity.Account, u *auth.User, lookupEmail string) *AdminUserDTO {
	email := acct.Email
	if email == "" {
		email = lookupEmail
	}
	return &AdminUserDTO{
		ID:          acct.ID,
		Username:    acct.Username,
		Email:       email,
		Role:        u.Role,
		CreatedAt:   acct.CreatedAt,
		LastSeenAt:  acct.LastSeenAt,
		Suspended:   u.IsSuspended(),
		SuspendedAt: u.SuspendedAt,
		SuspendTill: u.SuspendTill,
	}
}
