package application

import (
	"time"

	"jcourse/internal/domain/point"
)

type PointRecordDTO struct {
	Reason      point.RecordReason `json:"reason"`
	Amount      int                `json:"amount"`
	Description string             `json:"description"`
	CreatedAt   time.Time          `json:"created_at"`
}

type PointSummaryDTO struct {
	Total   int                             `json:"total"`
	Records PaginatedResult[PointRecordDTO] `json:"records"`
}

func newPointRecordDTO(r *point.Record) PointRecordDTO {
	return PointRecordDTO{
		Reason:      r.Reason,
		Amount:      r.Amount,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
	}
}
