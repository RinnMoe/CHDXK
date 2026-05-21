package announcement

import (
	"context"
	"time"
)

type AnnouncementView struct {
	ID        int
	Title     string
	Body      string
	Priority  int
	ShowStart time.Time
	ShowEnd   time.Time
	CreatedAt time.Time
}

type AnnouncementQuery interface {
	FindActive(ctx context.Context) ([]AnnouncementView, error)
}
