package repository

import (
	"context"
	"errors"
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

func newAnnouncementDomain(e *AnnouncementEntity) announcement.Announcement {
	return announcement.Announcement{
		ID:        e.ID,
		Title:     e.Title,
		Body:      e.Body,
		Priority:  e.Priority,
		ShowStart: e.ShowStart,
		ShowEnd:   e.ShowEnd,
		LinkURL:   e.LinkURL,
		LinkTitle: e.LinkTitle,
		CreatedAt: e.CreatedAt,
	}
}

func newAnnouncementView(e *AnnouncementEntity) announcement.AnnouncementView {
	return announcement.AnnouncementView{
		ID:        e.ID,
		Title:     e.Title,
		Body:      e.Body,
		Priority:  e.Priority,
		ShowStart: e.ShowStart,
		ShowEnd:   e.ShowEnd,
		LinkURL:   e.LinkURL,
		LinkTitle: e.LinkTitle,
		CreatedAt: e.CreatedAt,
	}
}

func newAnnouncementEntity(d *announcement.Announcement) AnnouncementEntity {
	return AnnouncementEntity{
		ID:        d.ID,
		Title:     d.Title,
		Body:      d.Body,
		Priority:  d.Priority,
		ShowStart: d.ShowStart,
		ShowEnd:   d.ShowEnd,
		LinkURL:   d.LinkURL,
		LinkTitle: d.LinkTitle,
		CreatedAt: d.CreatedAt,
	}
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
		result[i] = newAnnouncementView(&e)
	}
	return result, nil
}

func (r *AnnouncementRepository) FindAll(ctx context.Context) ([]announcement.AnnouncementView, error) {
	var entities []AnnouncementEntity
	if err := r.db.WithContext(ctx).
		Order("priority DESC, created_at DESC, id DESC").
		Find(&entities).Error; err != nil {
		return nil, err
	}

	result := make([]announcement.AnnouncementView, len(entities))
	for i, e := range entities {
		result[i] = newAnnouncementView(&e)
	}
	return result, nil
}

func (r *AnnouncementRepository) GetByID(ctx context.Context, id int) (*announcement.Announcement, error) {
	var e AnnouncementEntity
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := newAnnouncementDomain(&e)
	return &d, nil
}

func (r *AnnouncementRepository) Create(ctx context.Context, item *announcement.Announcement) error {
	e := newAnnouncementEntity(item)
	if err := r.db.WithContext(ctx).Create(&e).Error; err != nil {
		return err
	}
	item.ID = e.ID
	item.CreatedAt = e.CreatedAt
	return nil
}

func (r *AnnouncementRepository) Update(ctx context.Context, item *announcement.Announcement) error {
	e := newAnnouncementEntity(item)
	result := r.db.WithContext(ctx).
		Model(&AnnouncementEntity{}).
		Where("id = ?", item.ID).
		Updates(map[string]any{
			"title":      e.Title,
			"body":       e.Body,
			"priority":   e.Priority,
			"show_start": e.ShowStart,
			"show_end":   e.ShowEnd,
			"link_url":   e.LinkURL,
			"link_title": e.LinkTitle,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return announcement.ErrAnnouncementNotFound
	}
	return nil
}

func (r *AnnouncementRepository) Delete(ctx context.Context, id int) (bool, error) {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&AnnouncementEntity{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

var _ announcement.AnnouncementRepository = (*AnnouncementRepository)(nil)
