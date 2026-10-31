package canteen

import (
	"strings"
	"time"
	"unicode/utf8"

	"jcourse/pkg/apperr"
)

type ReportSubjectType string

const (
	ReportSubjectStore      ReportSubjectType = "store"
	ReportSubjectReview     ReportSubjectType = "review"
	ReportSubjectDishReview ReportSubjectType = "dish_review"
	ReportSubjectMedia      ReportSubjectType = "media"
)

func (t ReportSubjectType) Valid() bool {
	switch t {
	case ReportSubjectStore, ReportSubjectReview, ReportSubjectDishReview, ReportSubjectMedia:
		return true
	default:
		return false
	}
}

type ReportStatus string

const (
	ReportStatusOpen      ReportStatus = "open"
	ReportStatusReviewing ReportStatus = "reviewing"
	ReportStatusResolved  ReportStatus = "resolved"
	ReportStatusDismissed ReportStatus = "dismissed"
)

func (s ReportStatus) Valid() bool {
	switch s {
	case ReportStatusOpen, ReportStatusReviewing, ReportStatusResolved, ReportStatusDismissed:
		return true
	default:
		return false
	}
}

// ReportReason is a controlled vocabulary so the admin queue can aggregate.
type ReportReason string

const (
	ReportReasonSpam      ReportReason = "spam"
	ReportReasonAbuse     ReportReason = "abuse"
	ReportReasonFalseInfo ReportReason = "false_info"
	ReportReasonPrivacy   ReportReason = "privacy"
	ReportReasonOther     ReportReason = "other"
)

func (r ReportReason) Valid() bool {
	switch r {
	case ReportReasonSpam, ReportReasonAbuse, ReportReasonFalseInfo, ReportReasonPrivacy, ReportReasonOther:
		return true
	default:
		return false
	}
}

const MaxReportDetailRunes = 500

// ReportFilter drives the admin report queue.
type ReportFilter struct {
	Status   ReportStatus
	Page     int
	PageSize int
}

type Report struct {
	ID             int
	ReporterUserID int
	SubjectType    ReportSubjectType
	SubjectID      int
	ReasonCode     ReportReason
	Detail         string
	Status         ReportStatus
	HandledBy      *int
	HandledAt      *time.Time
	ResolutionNote string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (r *Report) ApplyDefaults() {
	r.Detail = strings.TrimSpace(r.Detail)
	if r.Status == "" {
		r.Status = ReportStatusOpen
	}
}

func (r *Report) Validate() error {
	if !r.SubjectType.Valid() || r.SubjectID <= 0 {
		return apperr.BadRequest("举报对象无效")
	}
	if !r.ReasonCode.Valid() {
		return ErrReportReason
	}
	if utf8.RuneCountInString(r.Detail) > MaxReportDetailRunes {
		return apperr.BadRequest("举报说明过长")
	}
	if !r.Status.Valid() {
		return apperr.BadRequest("举报状态无效")
	}
	return nil
}
