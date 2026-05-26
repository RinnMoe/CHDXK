package application

import (
	"context"
	"time"

	"jcourse/internal/domain/audit"
)

type AuditLogListFilter struct {
	StartTime   string `form:"start_time"`
	EndTime     string `form:"end_time"`
	Action      string `form:"action"`
	ActorUserID int    `form:"actor_user_id"`
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}

type AuditLogQueryService struct {
	query audit.QueryRepository
}

func NewAuditLogQueryService(query audit.QueryRepository) *AuditLogQueryService {
	return &AuditLogQueryService{query: query}
}

func (s *AuditLogQueryService) List(ctx context.Context, f AuditLogListFilter) (*PaginatedResult[AuditLogDTO], error) {
	startTime, _, err := parseOptionalAuditTime(f.StartTime)
	if err != nil {
		return nil, err
	}
	endTime, endDateOnly, err := parseOptionalAuditTime(f.EndTime)
	if err != nil {
		return nil, err
	}
	if endDateOnly {
		endTime = endTime.Add(24*time.Hour - time.Nanosecond)
	}
	if !startTime.IsZero() && !endTime.IsZero() && endTime.Before(startTime) {
		return nil, audit.ErrInvalidTimeRange
	}

	logs, total, err := s.query.Find(ctx, audit.LogFilter{
		StartTime:   startTime,
		EndTime:     endTime,
		Action:      f.Action,
		ActorUserID: f.ActorUserID,
		Page:        f.Page,
		PageSize:    f.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]AuditLogDTO, len(logs))
	for i := range logs {
		items[i] = newAuditLogDTO(logs[i])
	}
	return &PaginatedResult[AuditLogDTO]{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

func parseOptionalAuditTime(value string) (time.Time, bool, error) {
	if value == "" {
		return time.Time{}, false, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, false, nil
	}
	t, err := time.ParseInLocation("2006-01-02", value, time.Local)
	return t, true, err
}
