package application

import (
	"context"
	"time"

	"jcourse/internal/domain/audit"
)

type AuditLogCommandService struct {
	repo audit.CommandRepository
}

func NewAuditLogCommandService(repo audit.CommandRepository) *AuditLogCommandService {
	return &AuditLogCommandService{repo: repo}
}

func (s *AuditLogCommandService) Record(ctx context.Context, payload audit.RecordLogPayload) error {
	occurredAt := payload.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	details := payload.Details
	if details == nil {
		details = audit.Details{}
	}
	return s.repo.Create(ctx, &audit.Log{
		OccurredAt:  occurredAt,
		ActorUserID: payload.ActorUserID,
		Action:      payload.Action,
		TargetType:  payload.TargetType,
		TargetID:    payload.TargetID,
		Details:     details,
	})
}
