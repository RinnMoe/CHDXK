package point

import (
	"context"
	"time"
)

type Record struct {
	ID          int
	UserID      int
	Reason      string
	Amount      int
	Description string
	CreatedAt   time.Time
}

type RecordFilter struct {
	UserID   int
	Page     int
	PageSize int
}

type Query interface {
	SumByUser(ctx context.Context, userID int) (int, error)
	FindRecordsByUser(ctx context.Context, filter RecordFilter) ([]Record, int64, error)
}
