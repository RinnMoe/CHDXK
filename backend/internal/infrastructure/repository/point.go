package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/point"
)

type PointRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewPointRepository(db *gorm.DB, cache ...*redis.Client) *PointRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &PointRepository{db: db, cache: client}
}

func newPointRecordDomain(e *UserPointRecordEntity) point.Record {
	return point.Record{
		ID:          e.ID,
		UserID:      e.UserID,
		Reason:      point.RecordReason(e.Reason),
		Amount:      e.Amount,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
	}
}

func (r *PointRepository) SumByUser(ctx context.Context, userID int) (int, error) {
	key := cacheKey("point", userID, "sum")
	if cached, ok := cacheGetJSON[int](ctx, r.cache, key); ok {
		return *cached, nil
	}

	var total int64
	err := r.db.WithContext(ctx).
		Model(&UserPointRecordEntity{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	result := int(total)
	cacheSetJSON(ctx, r.cache, key, result)
	return result, nil
}

func (r *PointRepository) FindRecordsByUser(ctx context.Context, filter point.RecordFilter) ([]point.Record, int64, error) {
	db := r.db.WithContext(ctx).
		Model(&UserPointRecordEntity{}).
		Where("user_id = ?", filter.UserID)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true})
	if filter.Page > 0 && filter.PageSize > 0 {
		db = db.Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize)
	}

	var entities []UserPointRecordEntity
	if err := db.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	records := make([]point.Record, len(entities))
	for i, e := range entities {
		records[i] = newPointRecordDomain(&e)
	}
	return records, total, nil
}

var _ point.Query = (*PointRepository)(nil)
