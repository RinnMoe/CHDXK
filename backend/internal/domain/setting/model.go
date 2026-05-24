package setting

import (
	"context"
	"errors"
)

var ErrInvalidCurrentSemester = errors.New("invalid current semester")

type UserSettings struct {
	UserID          int
	CurrentSemester string
}

type Repository interface {
	GetByUserID(ctx context.Context, userID int) (*UserSettings, error)
	Save(ctx context.Context, settings *UserSettings) error
}
