package account

import (
	"context"
	"time"
)

type Account struct {
	ID       int
	Username string
	Password string
	Email    string

	CreatedAt  time.Time
	LastSeenAt time.Time
}

func NewRegisteredAccount(username, passwordHash string, now time.Time) *Account {
	return &Account{
		Username:   username,
		Email:      username,
		Password:   passwordHash,
		CreatedAt:  now,
		LastSeenAt: now,
	}
}

type AccountRepository interface {
	Create(ctx context.Context, u *Account) error
	Update(ctx context.Context, u *Account) error
	TouchLastSeen(ctx context.Context, userID int, at time.Time) error
	FindByID(ctx context.Context, id int) (*Account, error)
	FindByUsername(ctx context.Context, username string) (*Account, error)
	FindByEmail(ctx context.Context, email string) (*Account, error)
}
