package auth

import (
	"context"
	"time"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	ID       int
	Username string
	Role     string

	Password string
	Email    string

	CreatedAt  time.Time
	LastSeenAt time.Time

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

func NewRegisteredUser(username, passwordHash string, now time.Time) *User {
	return &User{
		Username:   username,
		Email:      username,
		Role:       RoleUser,
		Password:   passwordHash,
		CreatedAt:  now,
		LastSeenAt: now,
	}
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	TouchLastSeen(ctx context.Context, userID int, at time.Time) error
	FindByID(ctx context.Context, id int) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}
