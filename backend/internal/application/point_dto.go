package application

import (
	"time"

	"jcourse/internal/domain/point"
)

type PointRecordDTO struct {
	ID          int                `json:"id,omitempty"`
	ExternalID  string             `json:"external_id,omitempty"`
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
		ID:          r.ID,
		ExternalID:  r.ExternalID,
		Reason:      r.Reason,
		Amount:      r.Amount,
		Description: pointRecordDescription(r.Description),
		CreatedAt:   r.CreatedAt,
	}
}

func pointRecordDescription(description string) string {
	switch description {
	case "review_create":
		return "发布点评奖励"
	case "course_first_review":
		return "课程首评奖励"
	case "冲正：review_create":
		return "撤回发布点评奖励"
	case "冲正：course_first_review":
		return "撤回课程首评奖励"
	default:
		return description
	}
}
