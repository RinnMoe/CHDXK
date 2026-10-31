package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

// CanteenMediaRepository persists media metadata and review attachments.
type CanteenMediaRepository struct {
	db *gorm.DB
}

func NewCanteenMediaRepository(db *gorm.DB) *CanteenMediaRepository {
	return &CanteenMediaRepository{db: db}
}

func newCanteenMediaEntity(media *canteen.Media) CanteenMediaEntity {
	return CanteenMediaEntity{
		ID:            media.ID,
		OwnerUserID:   media.OwnerUserID,
		Purpose:       string(media.Purpose),
		StorageDriver: media.StorageDriver,
		StorageKey:    media.StorageKey,
		MIME:          media.MIME,
		ByteSize:      media.ByteSize,
		Width:         media.Width,
		Height:        media.Height,
		SHA256:        media.SHA256,
		Status:        string(media.Status),
		ReviewedBy:    media.ReviewedBy,
		ReviewedAt:    media.ReviewedAt,
		RejectReason:  media.RejectReason,
		ExpiresAt:     media.ExpiresAt,
		CreatedAt:     media.CreatedAt,
		UpdatedAt:     media.UpdatedAt,
	}
}

func (r *CanteenMediaRepository) Create(ctx context.Context, media *canteen.Media) error {
	entity := newCanteenMediaEntity(media)
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		if isCanteenUniqueViolation(err) {
			return apperr.Conflict("图片记录已存在")
		}
		return err
	}
	media.ID = entity.ID
	media.CreatedAt = entity.CreatedAt
	media.UpdatedAt = entity.UpdatedAt
	return nil
}

func (r *CanteenMediaRepository) Update(ctx context.Context, media *canteen.Media) error {
	result := r.db.WithContext(ctx).Model(&CanteenMediaEntity{}).Where("id = ?", media.ID).Updates(map[string]any{
		"mime":           media.MIME,
		"byte_size":      media.ByteSize,
		"width":          media.Width,
		"height":         media.Height,
		"sha256":         media.SHA256,
		"status":         string(media.Status),
		"reviewed_by":    media.ReviewedBy,
		"reviewed_at":    media.ReviewedAt,
		"reject_reason":  media.RejectReason,
		"expires_at":     media.ExpiresAt,
		"storage_key":    media.StorageKey,
		"storage_driver": media.StorageDriver,
		"updated_at":     media.UpdatedAt,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return canteen.ErrMediaNotFound
	}
	return nil
}

func (r *CanteenMediaRepository) ListByReview(ctx context.Context, reviewID int) ([]canteen.Media, error) {
	var entities []CanteenMediaEntity
	err := r.db.WithContext(ctx).
		Table("canteen_media AS m").
		Joins("JOIN canteen_review_media AS rm ON rm.media_id = m.id").
		Where("rm.review_event_id = ?", reviewID).
		Order("rm.sort_order ASC, m.id ASC").
		Select("m.*").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}
	media := make([]canteen.Media, 0, len(entities))
	for i := range entities {
		media = append(media, *newCanteenMediaDomain(&entities[i]))
	}
	return media, nil
}

func (r *CanteenMediaRepository) AttachToReview(ctx context.Context, reviewID, mediaID int, sortOrder int) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO canteen_review_media (review_event_id, media_id, sort_order)
		VALUES (?, ?, ?)
		ON CONFLICT DO NOTHING`, reviewID, mediaID, sortOrder).Error
}

func (r *CanteenMediaRepository) AttachToDish(ctx context.Context, dishID, mediaID, userID int, sortOrder int, now time.Time) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO canteen_dish_media (dish_id, media_id, user_id, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (dish_id, media_id) DO NOTHING`,
		dishID, mediaID, userID, sortOrder, now, now).Error
}

func (r *CanteenMediaRepository) CountByOwnerSince(ctx context.Context, userID int, since time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CanteenMediaEntity{}).
		Where("owner_user_id = ? AND created_at >= ?", userID, since).
		Count(&count).Error
	return int(count), err
}

func (r *CanteenMediaRepository) GetMedia(ctx context.Context, id int) (*canteen.Media, error) {
	var entity CanteenMediaEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return newCanteenMediaDomain(&entity), nil
}

func (r *CanteenMediaRepository) ListMediaByOwner(ctx context.Context, userID int, page, pageSize int) ([]canteen.Media, int64, error) {
	page, pageSize = normalizeCanteenPage(page, pageSize)
	query := r.db.WithContext(ctx).Model(&CanteenMediaEntity{}).Where("owner_user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var entities []CanteenMediaEntity
	if err := query.
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}
	media := make([]canteen.Media, 0, len(entities))
	for i := range entities {
		media = append(media, *newCanteenMediaDomain(&entities[i]))
	}
	return media, total, nil
}

func (r *CanteenMediaRepository) ListMediaByStatus(ctx context.Context, filter canteen.MediaFilter) ([]canteen.Media, int64, error) {
	page, pageSize := normalizeCanteenPage(filter.Page, filter.PageSize)
	query := r.db.WithContext(ctx).Model(&CanteenMediaEntity{})
	if filter.Status != "" {
		query = query.Where("status = ?", string(filter.Status))
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var entities []CanteenMediaEntity
	if err := query.
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&entities).Error; err != nil {
		return nil, 0, err
	}
	media := make([]canteen.Media, 0, len(entities))
	for i := range entities {
		media = append(media, *newCanteenMediaDomain(&entities[i]))
	}
	return media, total, nil
}
