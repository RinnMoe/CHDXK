package repository

import (
	"context"
	"maps"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/audit"
)

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func newAuditLogEntity(l *audit.Log) AuditLogEntity {
	details := datatypes.JSONMap{}
	maps.Copy(details, l.Details)
	return AuditLogEntity{
		ID:          l.ID,
		OccurredAt:  l.OccurredAt,
		ActorUserID: l.ActorUserID,
		Action:      l.Action,
		TargetType:  l.TargetType,
		TargetID:    l.TargetID,
		Details:     details,
		CreatedAt:   l.CreatedAt,
	}
}

func newAuditLogDomain(e *AuditLogEntity) audit.Log {
	details := audit.Details{}
	maps.Copy(details, e.Details)
	return audit.Log{
		ID:          e.ID,
		OccurredAt:  e.OccurredAt,
		ActorUserID: e.ActorUserID,
		Action:      e.Action,
		TargetType:  e.TargetType,
		TargetID:    e.TargetID,
		Details:     details,
		CreatedAt:   e.CreatedAt,
	}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *audit.Log) error {
	e := newAuditLogEntity(log)
	if err := gorm.G[AuditLogEntity](r.db).Create(ctx, &e); err != nil {
		return err
	}
	log.ID = e.ID
	log.CreatedAt = e.CreatedAt
	return nil
}

func (r *AuditLogRepository) Find(ctx context.Context, filter audit.LogFilter) ([]audit.Log, int64, error) {
	db := r.applyFilter(gorm.G[AuditLogEntity](r.db).Where("1 = 1"), filter)
	total, err := db.Count(ctx, "id")
	if err != nil {
		return nil, 0, err
	}

	db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: "occurred_at"}, Desc: true}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true})
	if filter.Page > 0 && filter.PageSize > 0 {
		db = db.Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize)
	}

	entities, err := db.Find(ctx)
	if err != nil {
		return nil, 0, err
	}
	logs := make([]audit.Log, len(entities))
	for i := range entities {
		logs[i] = newAuditLogDomain(&entities[i])
	}
	return logs, total, nil
}

func (r *AuditLogRepository) applyFilter(db gorm.ChainInterface[AuditLogEntity], filter audit.LogFilter) gorm.ChainInterface[AuditLogEntity] {
	if !filter.StartTime.IsZero() {
		db = db.Where("occurred_at >= ?", filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		db = db.Where("occurred_at <= ?", filter.EndTime)
	}
	if filter.Action != "" {
		db = db.Where("action = ?", filter.Action)
	}
	if filter.ActorUserID > 0 {
		db = db.Where("actor_user_id = ?", filter.ActorUserID)
	}
	return db
}

var _ audit.CommandRepository = (*AuditLogRepository)(nil)
var _ audit.QueryRepository = (*AuditLogRepository)(nil)
