package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/canteen"
)

// CanteenReportRepository persists user reports and admin resolutions.
type CanteenReportRepository struct {
	db *gorm.DB
}

func NewCanteenReportRepository(db *gorm.DB) *CanteenReportRepository {
	return &CanteenReportRepository{db: db}
}

func (r *CanteenReportRepository) Create(ctx context.Context, report *canteen.Report) error {
	entity := CanteenReportEntity{
		ID:             report.ID,
		ReporterUserID: report.ReporterUserID,
		SubjectType:    string(report.SubjectType),
		SubjectID:      report.SubjectID,
		ReasonCode:     string(report.ReasonCode),
		Detail:         report.Detail,
		Status:         string(report.Status),
		CreatedAt:      report.CreatedAt,
		UpdatedAt:      report.UpdatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		if isCanteenUniqueViolation(err) {
			return canteen.ErrReportDuplicate
		}
		return err
	}
	report.ID = entity.ID
	return nil
}

func (r *CanteenReportRepository) Get(ctx context.Context, id int) (*canteen.Report, error) {
	var entity CanteenReportEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return newCanteenReportDomain(&entity), nil
}

func (r *CanteenReportRepository) List(ctx context.Context, filter canteen.ReportFilter) ([]canteen.Report, int64, error) {
	page, pageSize := normalizeCanteenPage(filter.Page, filter.PageSize)
	query := r.db.WithContext(ctx).Model(&CanteenReportEntity{})
	if filter.Status != "" {
		query = query.Where("status = ?", string(filter.Status))
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var entities []CanteenReportEntity
	if err := query.
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}
	reports := make([]canteen.Report, 0, len(entities))
	for i := range entities {
		reports = append(reports, *newCanteenReportDomain(&entities[i]))
	}
	return reports, total, nil
}

func (r *CanteenReportRepository) Resolve(ctx context.Context, id int, handlerID int, status canteen.ReportStatus, note string, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&CanteenReportEntity{}).Where("id = ?", id).Updates(map[string]any{
		"status":          string(status),
		"handled_by":      handlerID,
		"handled_at":      now,
		"resolution_note": note,
		"updated_at":      now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return canteen.ErrReportNotFound
	}
	return nil
}
