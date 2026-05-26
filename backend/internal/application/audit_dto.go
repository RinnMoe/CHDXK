package application

import (
	"strconv"
	"time"

	"jcourse/internal/domain/audit"
)

type AuditLogDTO struct {
	ID          string        `json:"id"`
	OccurredAt  time.Time     `json:"occurred_at"`
	ActorUserID int           `json:"actor_user_id"`
	Action      string        `json:"action"`
	TargetType  string        `json:"target_type"`
	TargetID    string        `json:"target_id"`
	Details     audit.Details `json:"details"`
	CreatedAt   time.Time     `json:"created_at"`
}

func newAuditLogDTO(l audit.Log) AuditLogDTO {
	details := l.Details
	if details == nil {
		details = audit.Details{}
	}
	return AuditLogDTO{
		ID:          strconv.FormatInt(l.ID, 10),
		OccurredAt:  l.OccurredAt,
		ActorUserID: l.ActorUserID,
		Action:      l.Action,
		TargetType:  l.TargetType,
		TargetID:    l.TargetID,
		Details:     details,
		CreatedAt:   l.CreatedAt,
	}
}
