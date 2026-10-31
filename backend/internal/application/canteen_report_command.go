package application

import (
	"context"
	"strconv"
	"time"

	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

// CreateCanteenReportCommand is a user report submission.
type CreateCanteenReportCommand struct {
	SubjectType canteen.ReportSubjectType
	SubjectID   int
	Reason      canteen.ReportReason
	Detail      string
}

// CanteenReportCommandService handles reports and admin resolutions.
type CanteenReportCommandService struct {
	reports  canteen.ReportRepository
	settings CanteenSettingsProvider
	now      func() time.Time
}

func NewCanteenReportCommandService(reports canteen.ReportRepository, settings CanteenSettingsProvider) *CanteenReportCommandService {
	return &CanteenReportCommandService{reports: reports, settings: settings, now: time.Now}
}

func (s *CanteenReportCommandService) Create(ctx context.Context, user *auth.User, cmd CreateCanteenReportCommand) (*canteen.Report, error) {
	if s.settings != nil {
		config, err := s.settings.CanteenRuntimeConfig(ctx)
		if err != nil {
			return nil, err
		}
		if !config.Enabled {
			return nil, apperr.NotFound("食堂评价暂未开放")
		}
	}
	now := s.now()
	report := &canteen.Report{
		ReporterUserID: user.ID,
		SubjectType:    cmd.SubjectType,
		SubjectID:      cmd.SubjectID,
		ReasonCode:     cmd.Reason,
		Detail:         cmd.Detail,
		Status:         canteen.ReportStatusOpen,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	report.ApplyDefaults()
	if err := report.Validate(); err != nil {
		return nil, err
	}
	if err := s.reports.Create(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (s *CanteenReportCommandService) List(ctx context.Context, user *auth.User, filter canteen.ReportFilter) ([]CanteenReportDTO, int64, int, int, error) {
	if user == nil || !user.IsAdmin() {
		return nil, 0, 0, 0, canteen.ErrCannotModerate
	}
	page, pageSize := normalizeCanteenPagination(filter.Page, filter.PageSize)
	filter.Page, filter.PageSize = page, pageSize
	items, total, err := s.reports.List(ctx, filter)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	result := make([]CanteenReportDTO, 0, len(items))
	for _, item := range items {
		result = append(result, newCanteenReportDTO(item))
	}
	return result, total, page, pageSize, nil
}

func (s *CanteenReportCommandService) Resolve(ctx context.Context, user *auth.User, reportID int, status canteen.ReportStatus, note string) error {
	if user == nil || !user.IsAdmin() {
		return canteen.ErrCannotModerate
	}
	if status != canteen.ReportStatusResolved && status != canteen.ReportStatusDismissed {
		return apperr.BadRequest("举报处理状态无效")
	}
	now := s.now()
	if err := s.reports.Resolve(ctx, reportID, user.ID, status, note, now); err != nil {
		return err
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  now,
		ActorUserID: user.ID,
		Action:      "canteen.report.resolve",
		TargetType:  "canteen_report",
		TargetID:    strconv.Itoa(reportID),
		Details:     audit.Details{"status": string(status)},
	})
	return nil
}
