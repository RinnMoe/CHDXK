package setting

import (
	"context"

	"jcourse/pkg/apperr"
)

var ErrInvalidCurrentSemester = apperr.ErrInvalidCurrentSemester

type UserSettings struct {
	UserID          int
	CurrentSemester string
}

type Repository interface {
	GetByUserID(ctx context.Context, userID int) (*UserSettings, error)
	Save(ctx context.Context, settings *UserSettings) error
}
