package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/announcement"
)

type AnnouncementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

func (r *AnnouncementRepository) FindActive(ctx context.Context) ([]announcement.AnnouncementView, error) {
	now := time.Now()
	var entities []AnnouncementEntity
	if err := r.db.WithContext(ctx).
		Where("show_start <= ? AND show_end >= ?", now, now).
		Order("priority DESC, created_at DESC").
		Find(&entities).Error; err != nil {
		return nil, err
	}

	result := make([]announcement.AnnouncementView, len(entities))
	for i, e := range entities {
		result[i] = announcement.AnnouncementView{
			ID:        e.ID,
			Title:     e.Title,
			Body:      e.Body,
			Priority:  e.Priority,
			ShowStart: e.ShowStart,
			ShowEnd:   e.ShowEnd,
			CreatedAt: e.CreatedAt,
		}
	}
	return result, nil
}

var _ announcement.AnnouncementQuery = (*AnnouncementRepository)(nil)
