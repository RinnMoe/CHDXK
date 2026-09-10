package auth

import (
	"context"
	"time"
)

const (
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
	RoleSystem     = "system"
	RoleUser       = "user"
)

type User struct {
	ID               int
	Role             string
	Subject          string
	Username         string
	Email            string
	IdentityRole     string
	IdentityType     string
	PermissionMode   string
	ViolationCount   int
	InitialPassword  bool
	HasPointsAccount bool
	SessionID        string
	ExpiresAt        time.Time
	ImpersonatedBy   string
	SuspendedAt      *time.Time
	SuspendTill      *time.Time
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
	return u.Role == RoleAdmin || u.IsSuperAdmin()
}

func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

func (u *User) IsSystemAPIKey() bool {
	return u.Role == RoleSystem
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
	FindAdmin(ctx context.Context) ([]User, error)
}
