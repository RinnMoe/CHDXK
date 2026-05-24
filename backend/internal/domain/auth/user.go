package auth

import (
	"context"
	"time"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type AdminConfig struct {
	DefaultSuspendDays int
}

func DefaultAdminConfig() AdminConfig {
	return AdminConfig{DefaultSuspendDays: 30}
}

type User struct {
	ID          int
	Role        string
	SuspendedAt *time.Time
	SuspendTill *time.Time
}

func (u *User) IsSuspended() bool {
	if u.SuspendTill == nil {
		return u.SuspendedAt != nil
	}
	return u.SuspendTill.After(time.Now())
}

func (u *User) SuspensionExpired() bool {
	return u.SuspendTill != nil && !u.SuspendTill.After(time.Now())
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) Suspend(d time.Duration) {
	now := time.Now()
	u.SuspendTill = new(now.Add(d))
	u.SuspendedAt = &now
}

func (u *User) ClearSuspension() {
	u.SuspendedAt = nil
	u.SuspendTill = nil
}

type UserRepository interface {
	Update(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id int) (*User, error)
	FindByRole(ctx context.Context, role string) ([]User, error)
}
