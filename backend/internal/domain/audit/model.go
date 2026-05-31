package audit

import (
	"context"
	"time"

	"jcourse/pkg/apperr"
)

const (
	ActionUserSuspend               = "user.suspend"
	ActionUserUnsuspend             = "user.unsuspend"
	ActionUserPasswordReset         = "user.password_reset"
	ActionAdminGrant                = "admin.grant"
	ActionAdminRevoke               = "admin.revoke"
	ActionSystemAPIKeyCreate        = "system_api_key.create"
	ActionSystemAPIKeyDelete        = "system_api_key.delete"
	ActionReviewUpdate              = "review.update"
	ActionReviewDelete              = "review.delete"
	ActionReviewModeratorRemarkEdit = "review.moderator_remark.update"

	TargetTypeUser         = "user"
	TargetTypeSystemAPIKey = "system_api_key"
	TargetTypeReview       = "review"
)

var ErrInvalidTimeRange = apperr.ErrInvalidAuditTimeRange

type Details map[string]any

type Log struct {
	ID          int64
	OccurredAt  time.Time
	ActorUserID int
	Action      string
	TargetType  string
	TargetID    string
	Details     Details
	CreatedAt   time.Time
}

type LogFilter struct {
	StartTime   time.Time
	EndTime     time.Time
	Action      string
	ActorUserID int
	Page        int
	PageSize    int
}

type CommandRepository interface {
	Create(ctx context.Context, log *Log) error
}

type QueryRepository interface {
	Find(ctx context.Context, filter LogFilter) ([]Log, int64, error)
}
