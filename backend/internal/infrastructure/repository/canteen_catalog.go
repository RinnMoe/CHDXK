package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

func isCanteenUniqueViolation(err error) bool {
	var sqlError interface{ SQLState() string }
	return errors.As(err, &sqlError) && sqlError.SQLState() == "23505"
}

// CanteenHallRepository persists canteen buildings.
type CanteenHallRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewCanteenHallRepository(db *gorm.DB, cache ...*redis.Client) *CanteenHallRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &CanteenHallRepository{db: db, cache: client}
}

func (r *CanteenHallRepository) Create(ctx context.Context, hall *canteen.Hall) error {
	entity := newCanteenHallEntity(hall)
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		if isCanteenUniqueViolation(err) {
			return apperr.Conflict("该校区已存在同名食堂")
		}
		return err
	}
	hall.ID = entity.ID
	hall.CreatedAt = entity.CreatedAt
	hall.UpdatedAt = entity.UpdatedAt
	r.invalidate(ctx)
	return nil
}

func (r *CanteenHallRepository) Update(ctx context.Context, hall *canteen.Hall) error {
	result := r.db.WithContext(ctx).Model(&CanteenHallEntity{}).Where("id = ?", hall.ID).Updates(map[string]any{
		"campus_key":        hall.CampusKey,
		"name":              hall.Name,
		"short_name":        hall.ShortName,
		"address":           hall.Address,
		"latitude":          hall.Latitude,
		"longitude":         hall.Longitude,
		"external_place_id": hall.ExternalPlaceID,
		"status":            string(hall.Status),
		"display_order":     hall.DisplayOrder,
		"updated_at":        hall.UpdatedAt,
	})
	if result.Error != nil {
		if isCanteenUniqueViolation(result.Error) {
			return apperr.Conflict("该校区已存在同名食堂")
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return canteen.ErrHallNotFound
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenHallRepository) Get(ctx context.Context, id int) (*canteen.Hall, error) {
	var entity CanteenHallEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return newCanteenHallDomain(&entity), nil
}

func (r *CanteenHallRepository) invalidate(ctx context.Context) {
	cacheDeletePattern(ctx, r.cache, cacheKey("canteen", "*"))
}

// CanteenStoreRepository persists stores and their persisted aggregate stats.
type CanteenStoreRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewCanteenStoreRepository(db *gorm.DB, cache ...*redis.Client) *CanteenStoreRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &CanteenStoreRepository{db: db, cache: client}
}

func (r *CanteenStoreRepository) Create(ctx context.Context, store *canteen.Store) error {
	entity := newCanteenStoreEntity(store)
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		if isCanteenUniqueViolation(err) {
			return apperr.Conflict("该食堂下已存在同名店铺")
		}
		return err
	}
	store.ID = entity.ID
	store.CreatedAt = entity.CreatedAt
	store.UpdatedAt = entity.UpdatedAt
	r.invalidate(ctx)
	return nil
}

func (r *CanteenStoreRepository) Update(ctx context.Context, store *canteen.Store) error {
	result := r.db.WithContext(ctx).Model(&CanteenStoreEntity{}).Where("id = ?", store.ID).Updates(map[string]any{
		"hall_id":              store.HallID,
		"name":                 store.Name,
		"normalized_name":      store.NormalizedName,
		"floor":                store.Floor,
		"categories":           canteenStringArray(store.Categories),
		"status":               string(store.Status),
		"merged_into_store_id": store.MergedIntoStoreID,
		"cover_media_id":       store.CoverMediaID,
		"updated_at":           store.UpdatedAt,
	})
	if result.Error != nil {
		if isCanteenUniqueViolation(result.Error) {
			return apperr.Conflict("该食堂下已存在同名店铺")
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return canteen.ErrStoreNotFound
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenStoreRepository) Get(ctx context.Context, id int) (*canteen.Store, error) {
	var entity CanteenStoreEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return newCanteenStoreDomain(&entity), nil
}

func (r *CanteenStoreRepository) GetByHallAndNormalizedName(ctx context.Context, hallID int, normalizedName string) (*canteen.Store, error) {
	var entity CanteenStoreEntity
	err := r.db.WithContext(ctx).
		Where("hall_id = ? AND normalized_name = ? AND status <> ?", hallID, normalizedName, string(canteen.StoreStatusMerged)).
		Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return newCanteenStoreDomain(&entity), nil
}

func (r *CanteenStoreRepository) UpdateStatus(ctx context.Context, id int, status canteen.StoreStatus, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&CanteenStoreEntity{}).Where("id = ?", id).Updates(map[string]any{
		"status":     string(status),
		"updated_at": now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return canteen.ErrStoreNotFound
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenStoreRepository) Merge(ctx context.Context, id, targetID int, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&CanteenStoreEntity{}).Where("id = ?", id).Updates(map[string]any{
		"status":               string(canteen.StoreStatusMerged),
		"merged_into_store_id": targetID,
		"updated_at":           now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return canteen.ErrStoreNotFound
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenStoreRepository) UpdateStats(ctx context.Context, id int, stats canteen.StoreStats, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&CanteenStoreEntity{}).Where("id = ?", id).Updates(map[string]any{
		"rating_count":       stats.RatingCount,
		"rating_avg":         stats.RatingAvg,
		"rating_score":       stats.RatingScore,
		"taste_avg":          stats.TasteAvg,
		"value_avg":          stats.ValueAvg,
		"queue_avg":          stats.QueueAvg,
		"hygiene_avg":        stats.HygieneAvg,
		"service_avg":        stats.ServiceAvg,
		"review_event_count": stats.ReviewEventCount,
		"last_review_at":     stats.LastReviewAt,
		"rating_updated_at":  stats.RatingUpdatedAt,
		"updated_at":         now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return canteen.ErrStoreNotFound
	}
	r.invalidate(ctx)
	return nil
}

// CanteenDishRepository persists dishes and their persisted aggregate stats.
type CanteenDishRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewCanteenDishRepository(db *gorm.DB, cache ...*redis.Client) *CanteenDishRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &CanteenDishRepository{db: db, cache: client}
}

func (r *CanteenDishRepository) Create(ctx context.Context, dish *canteen.Dish) error {
	entity := newCanteenDishEntity(dish)
	if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
		if isCanteenUniqueViolation(err) {
			return apperr.Conflict("该店铺下已存在同名菜品")
		}
		return err
	}
	dish.ID = entity.ID
	dish.CreatedAt = entity.CreatedAt
	dish.UpdatedAt = entity.UpdatedAt
	r.invalidate(ctx)
	return nil
}

func (r *CanteenDishRepository) Update(ctx context.Context, dish *canteen.Dish) error {
	result := r.db.WithContext(ctx).Model(&CanteenDishEntity{}).Where("id = ?", dish.ID).Updates(map[string]any{
		"store_id":            dish.StoreID,
		"name":                dish.Name,
		"normalized_name":     dish.NormalizedName,
		"status":              string(dish.Status),
		"merged_into_dish_id": dish.MergedIntoDishID,
		"updated_at":          dish.UpdatedAt,
	})
	if result.Error != nil {
		if isCanteenUniqueViolation(result.Error) {
			return apperr.Conflict("该店铺下已存在同名菜品")
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return canteen.ErrDishNotFound
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenDishRepository) Get(ctx context.Context, id int) (*canteen.Dish, error) {
	var entity CanteenDishEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return newCanteenDishDomain(&entity), nil
}

func (r *CanteenDishRepository) GetByStoreAndNormalizedName(ctx context.Context, storeID int, normalizedName string) (*canteen.Dish, error) {
	var entity CanteenDishEntity
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND normalized_name = ? AND status <> ?", storeID, normalizedName, string(canteen.DishStatusMerged)).
		Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return newCanteenDishDomain(&entity), nil
}

func (r *CanteenDishRepository) UpdateStatus(ctx context.Context, id int, status canteen.DishStatus, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&CanteenDishEntity{}).Where("id = ?", id).Updates(map[string]any{
		"status":     string(status),
		"updated_at": now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return canteen.ErrDishNotFound
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenDishRepository) Merge(ctx context.Context, id, targetID int, now time.Time) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var source CanteenDishEntity
		if err := tx.First(&source, id).Error; err != nil {
			return err
		}
		result := tx.Model(&CanteenDishEntity{}).Where("id = ?", id).Updates(map[string]any{
			"status":              string(canteen.DishStatusMerged),
			"merged_into_dish_id": targetID,
			"updated_at":          now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return canteen.ErrDishNotFound
		}
		if err := tx.Exec(
			"UPDATE canteen_dish_review_events SET dish_id = ? WHERE dish_id = ?",
			targetID, id,
		).Error; err != nil {
			return err
		}
		// Preserve synonyms so future searches still resolve to the target.
		if err := tx.Exec(`
			INSERT INTO canteen_dish_aliases (dish_id, name, normalized_name, created_at)
			SELECT ?, name, normalized_name, ? FROM canteen_dish_aliases WHERE dish_id = ?
			ON CONFLICT (dish_id, normalized_name) DO NOTHING`,
			targetID, now, id,
		).Error; err != nil {
			return err
		}
		if source.NormalizedName != "" {
			if err := tx.Exec(`
				INSERT INTO canteen_dish_aliases (dish_id, name, normalized_name, created_at)
				VALUES (?, ?, ?, ?)
				ON CONFLICT (dish_id, normalized_name) DO NOTHING`,
				targetID, source.Name, source.NormalizedName, now,
			).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec("DELETE FROM canteen_dish_aliases WHERE dish_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			INSERT INTO canteen_dish_media (dish_id, media_id, user_id, sort_order, created_at, updated_at)
			SELECT ?, media_id, user_id, sort_order, created_at, updated_at
			FROM canteen_dish_media WHERE dish_id = ?
			ON CONFLICT (dish_id, media_id) DO NOTHING`,
			targetID, id,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM canteen_dish_media WHERE dish_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			"DELETE FROM canteen_dish_rating_summaries WHERE dish_id = ?", id,
		).Error; err != nil {
			return err
		}
		var userIDs []int
		if err := tx.Raw(
			"SELECT DISTINCT user_id FROM canteen_dish_review_events WHERE dish_id = ?", targetID,
		).Scan(&userIDs).Error; err != nil {
			return err
		}
		for _, userID := range userIDs {
			if err := recomputeCanteenDishSummary(tx, targetID, userID); err != nil {
				return err
			}
		}
		if err := recomputeCanteenDishStats(tx, targetID, canteen.DefaultPriorCount); err != nil {
			return err
		}
		return recomputeCanteenDishStats(tx, id, canteen.DefaultPriorCount)
	})
	if err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenDishRepository) UpdateStats(ctx context.Context, id int, stats canteen.DishStats, now time.Time) error {
	result := r.db.WithContext(ctx).Model(&CanteenDishEntity{}).Where("id = ?", id).Updates(map[string]any{
		"rating_count":       stats.RatingCount,
		"rating_avg":         stats.RatingAvg,
		"rating_score":       stats.RatingScore,
		"taste_avg":          stats.TasteAvg,
		"portion_avg":        stats.PortionAvg,
		"value_avg":          stats.ValueAvg,
		"review_event_count": stats.ReviewEventCount,
		"rating_updated_at":  stats.RatingUpdatedAt,
		"updated_at":         now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return canteen.ErrDishNotFound
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenStoreRepository) invalidate(ctx context.Context) {
	cacheDeletePattern(ctx, r.cache, cacheKey("canteen", "*"))
}

func (r *CanteenDishRepository) invalidate(ctx context.Context) {
	cacheDeletePattern(ctx, r.cache, cacheKey("canteen", "*"))
}
