package identity

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAlreadyExists = errors.New("user already exists")
	ErrNotFound      = errors.New("user not found")
)

type Account struct {
	ID           int
	Username     string
	PasswordHash string
	Email        string

	CreatedAt  time.Time
	LastSeenAt time.Time
}

func NewRegisteredAccount(username, passwordHash string, now time.Time) *Account {
	return &Account{
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		LastSeenAt:   now,
	}
}

type Repository interface {
	Create(ctx context.Context, account *Account) error
	Update(ctx context.Context, account *Account) error
	FindByID(ctx context.Context, id int) (*Account, error)
	FindByUsername(ctx context.Context, username string) (*Account, error)
	FindByEmail(ctx context.Context, email string) (*Account, error)
}
