package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"jcourse/internal/domain/canteen"
)

type canteenDishRow struct {
	ID               int     `gorm:"column:id"`
	StoreID          int     `gorm:"column:store_id"`
	StoreName        string  `gorm:"column:store_name"`
	Name             string  `gorm:"column:name"`
	Status           string  `gorm:"column:status"`
	RatingCount      int     `gorm:"column:rating_count"`
	RatingAvg        float64 `gorm:"column:rating_avg"`
	RatingScore      float64 `gorm:"column:rating_score"`
	TasteAvg         float64 `gorm:"column:taste_avg"`
	PortionAvg       float64 `gorm:"column:portion_avg"`
	ValueAvg         float64 `gorm:"column:value_avg"`
	ReviewEventCount int     `gorm:"column:review_event_count"`
	CoverMediaID     int     `gorm:"column:cover_media_id"`
	PhotoCount       int     `gorm:"column:photo_count"`
}

type canteenDishReviewRow struct {
	ID           int       `gorm:"column:id"`
	DishID       int       `gorm:"column:dish_id"`
	DishName     string    `gorm:"column:dish_name"`
	StoreID      int       `gorm:"column:store_id"`
	StoreName    string    `gorm:"column:store_name"`
	UserID       int       `gorm:"column:user_id"`
	Taste        int       `gorm:"column:taste"`
	Portion      int       `gorm:"column:portion"`
	Value        int       `gorm:"column:value"`
	OverallScore float64   `gorm:"column:overall_score"`
	Content      string    `gorm:"column:content"`
	Status       string    `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func newCanteenDishView(row canteenDishRow) canteen.DishView {
	view := canteen.DishView{
		ID:               row.ID,
		StoreID:          row.StoreID,
		StoreName:        row.StoreName,
		Name:             row.Name,
		Status:           canteen.DishStatus(row.Status),
		RatingCount:      row.RatingCount,
		RatingAvg:        row.RatingAvg,
		RatingScore:      row.RatingScore,
		TasteAvg:         row.TasteAvg,
		PortionAvg:       row.PortionAvg,
		ValueAvg:         row.ValueAvg,
		ReviewEventCount: row.ReviewEventCount,
		PhotoCount:       row.PhotoCount,
	}
	if row.CoverMediaID > 0 {
		view.CoverURL = fmt.Sprintf("/api/canteen/media/%d", row.CoverMediaID)
	}
	return view
}

// CanteenDishQueryRepository reads the dish directory and a user's dish ratings.
type CanteenDishQueryRepository struct {
	db *gorm.DB
}

func NewCanteenDishQueryRepository(db *gorm.DB) *CanteenDishQueryRepository {
	return &CanteenDishQueryRepository{db: db}
}

const canteenDishSelect = `
	d.id AS id,
	d.store_id AS store_id,
	s.name AS store_name,
	d.name AS name,
	d.status AS status,
	d.rating_count AS rating_count,
	d.rating_avg AS rating_avg,
	d.rating_score AS rating_score,
	d.taste_avg AS taste_avg,
	d.portion_avg AS portion_avg,
	d.value_avg AS value_avg,
	d.review_event_count AS review_event_count,
	COALESCE((
		SELECT m.id FROM canteen_dish_media dm
		JOIN canteen_media m ON m.id = dm.media_id
		WHERE dm.dish_id = d.id AND m.status = 'approved'
		ORDER BY dm.sort_order ASC, m.id ASC
		LIMIT 1
	), 0) AS cover_media_id,
	(
		SELECT count(*) FROM canteen_dish_media dm
		JOIN canteen_media m ON m.id = dm.media_id
		WHERE dm.dish_id = d.id AND m.status = 'approved'
	) AS photo_count`

func (r *CanteenDishQueryRepository) ListDishes(ctx context.Context, filter canteen.DishListFilter) ([]canteen.DishView, error) {
	query := r.db.WithContext(ctx).
		Table("canteen_dishes AS d").
		Joins("JOIN canteen_stores AS s ON s.id = d.store_id")
	if filter.StoreID > 0 {
		query = query.Where("d.store_id = ?", filter.StoreID)
	}
	status := filter.Status
	if status == "" {
		status = canteen.DishStatusActive
	}
	query = query.Where("d.status = ?", string(status))
	if term := canteen.NormalizeName(filter.Q); term != "" {
		pattern := "%" + term + "%"
		query = query.Where(`(
			d.normalized_name LIKE ? OR EXISTS (
				SELECT 1 FROM canteen_dish_aliases a
				WHERE a.dish_id = d.id AND a.normalized_name LIKE ?
			)
		)`, pattern, pattern)
	}
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows []canteenDishRow
	if err := query.
		Select(canteenDishSelect).
		Order("d.rating_score DESC, d.rating_count DESC, d.name ASC, d.id ASC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	dishes := make([]canteen.DishView, 0, len(rows))
	for _, row := range rows {
		dishes = append(dishes, newCanteenDishView(row))
	}
	return dishes, nil
}

func (r *CanteenDishQueryRepository) GetDish(ctx context.Context, id int) (*canteen.Dish, error) {
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

func (r *CanteenDishQueryRepository) ListDishPhotos(ctx context.Context, dishID int) ([]canteen.MediaView, error) {
	type photoRow struct {
		MediaID int `gorm:"column:media_id"`
		Width   int `gorm:"column:width"`
		Height  int `gorm:"column:height"`
	}
	var rows []photoRow
	if err := r.db.WithContext(ctx).
		Table("canteen_dish_media AS dm").
		Joins("JOIN canteen_media AS m ON m.id = dm.media_id").
		Where("dm.dish_id = ? AND m.status = ?", dishID, string(canteen.MediaStatusApproved)).
		Order("dm.sort_order ASC, m.id ASC").
		Select("dm.media_id AS media_id, m.width AS width, m.height AS height").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	photos := make([]canteen.MediaView, 0, len(rows))
	for _, row := range rows {
		photos = append(photos, canteen.MediaView{
			ID:     row.MediaID,
			Width:  row.Width,
			Height: row.Height,
			URL:    fmt.Sprintf("/api/canteen/media/%d", row.MediaID),
		})
	}
	return photos, nil
}

func (r *CanteenDishQueryRepository) ListDishReviews(ctx context.Context, dishID int, page, pageSize int) ([]canteen.DishReviewView, int64, error) {
	page, pageSize = normalizeCanteenPage(page, pageSize)
	query := r.db.WithContext(ctx).
		Table("canteen_dish_review_events AS r").
		Joins("JOIN canteen_dishes AS d ON d.id = r.dish_id").
		Joins("JOIN canteen_stores AS s ON s.id = d.store_id").
		Where("r.dish_id = ? AND r.status = ?", dishID, string(canteen.ReviewStatusActive))
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []canteenDishReviewRow
	if err := query.
		Select(`
			r.id AS id,
			r.dish_id AS dish_id,
			d.name AS dish_name,
			d.store_id AS store_id,
			s.name AS store_name,
			r.user_id AS user_id,
			r.taste AS taste,
			r.portion AS portion,
			r.value AS value,
			r.overall_score AS overall_score,
			r.content AS content,
			r.status AS status,
			r.created_at AS created_at`).
		Order("r.created_at DESC, r.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	views := make([]canteen.DishReviewView, 0, len(rows))
	for _, row := range rows {
		views = append(views, canteen.DishReviewView{
			ID:           row.ID,
			DishID:       row.DishID,
			DishName:     row.DishName,
			StoreID:      row.StoreID,
			StoreName:    row.StoreName,
			UserID:       row.UserID,
			Taste:        row.Taste,
			Portion:      row.Portion,
			Value:        row.Value,
			OverallScore: row.OverallScore,
			Content:      row.Content,
			Status:       canteen.ReviewStatus(row.Status),
			CreatedAt:    row.CreatedAt,
		})
	}
	return views, total, nil
}

func (r *CanteenDishQueryRepository) ListDishReviewsByUser(ctx context.Context, userID int, page, pageSize int) ([]canteen.DishReviewView, int64, error) {
	page, pageSize = normalizeCanteenPage(page, pageSize)
	query := r.db.WithContext(ctx).
		Table("canteen_dish_review_events AS r").
		Joins("JOIN canteen_dishes AS d ON d.id = r.dish_id").
		Joins("JOIN canteen_stores AS s ON s.id = d.store_id").
		Where("r.user_id = ? AND r.status <> ?", userID, string(canteen.ReviewStatusDeleted))
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []canteenDishReviewRow
	if err := query.
		Select(`
			r.id AS id,
			r.dish_id AS dish_id,
			d.name AS dish_name,
			d.store_id AS store_id,
			s.name AS store_name,
			r.user_id AS user_id,
			r.taste AS taste,
			r.portion AS portion,
			r.value AS value,
			r.overall_score AS overall_score,
			r.content AS content,
			r.status AS status,
			r.created_at AS created_at`).
		Order("r.created_at DESC, r.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	views := make([]canteen.DishReviewView, 0, len(rows))
	for _, row := range rows {
		views = append(views, canteen.DishReviewView{
			ID:           row.ID,
			DishID:       row.DishID,
			DishName:     row.DishName,
			StoreID:      row.StoreID,
			StoreName:    row.StoreName,
			UserID:       row.UserID,
			Taste:        row.Taste,
			Portion:      row.Portion,
			Value:        row.Value,
			OverallScore: row.OverallScore,
			Content:      row.Content,
			Status:       canteen.ReviewStatus(row.Status),
			CreatedAt:    row.CreatedAt,
		})
	}
	return views, total, nil
}

// CanteenDishReviewRepository persists dish ratings and recomputes dish stats.
type CanteenDishReviewRepository struct {
	db         *gorm.DB
	cache      *redis.Client
	priorCount int
}

func NewCanteenDishReviewRepository(db *gorm.DB, cache ...*redis.Client) *CanteenDishReviewRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &CanteenDishReviewRepository{db: db, cache: client, priorCount: canteen.DefaultPriorCount}
}

func (r *CanteenDishReviewRepository) SetPriorCount(prior int) {
	if prior > 0 {
		r.priorCount = prior
	}
}

func (r *CanteenDishReviewRepository) Create(ctx context.Context, review *canteen.DishReviewEvent) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dishID int
		if err := tx.Raw(
			"SELECT id FROM canteen_dishes WHERE id = ? AND status = ? FOR UPDATE",
			review.DishID, string(canteen.DishStatusActive),
		).Scan(&dishID).Error; err != nil {
			return err
		}
		if dishID == 0 {
			return canteen.ErrDishNotFound
		}
		entity := CanteenDishReviewEventEntity{
			DishID:        review.DishID,
			UserID:        review.UserID,
			ReviewEventID: review.ReviewEventID,
			Taste:         review.DishDimensions.Taste,
			Portion:       review.DishDimensions.Portion,
			Value:         review.DishDimensions.Value,
			OverallScore:  review.OverallScore,
			Content:       review.Content,
			Status:        string(review.Status),
			CreatedAt:     review.CreatedAt,
			UpdatedAt:     review.UpdatedAt,
		}
		if err := tx.Create(&entity).Error; err != nil {
			return err
		}
		review.ID = entity.ID
		if err := recomputeCanteenDishSummary(tx, review.DishID, review.UserID); err != nil {
			return err
		}
		return recomputeCanteenDishStats(tx, review.DishID, r.priorCount)
	})
	if err != nil {
		return err
	}
	cacheDeletePattern(ctx, r.cache, cacheKey("canteen", "*"))
	return nil
}

func recomputeCanteenDishSummary(tx *gorm.DB, dishID, userID int) error {
	var activeCount int
	if err := tx.Raw(`
		SELECT count(*) FROM canteen_dish_review_events
		WHERE dish_id = ? AND user_id = ? AND status = ?`,
		dishID, userID, string(canteen.ReviewStatusActive)).Scan(&activeCount).Error; err != nil {
		return err
	}
	if activeCount == 0 {
		return tx.Exec(
			"DELETE FROM canteen_dish_rating_summaries WHERE dish_id = ? AND user_id = ?",
			dishID, userID,
		).Error
	}
	return tx.Exec(`
		WITH recent AS (
			SELECT overall_score, created_at, id,
			       row_number() OVER (ORDER BY created_at DESC, id DESC) AS rn
			FROM canteen_dish_review_events
			WHERE dish_id = ? AND user_id = ? AND status = ?
			ORDER BY created_at DESC, id DESC
			LIMIT 3
		), weighted AS (
			SELECT sum(overall_score * w) / sum(w) AS score, count(*) AS window_size
			FROM (
				SELECT overall_score,
				       CASE rn WHEN 1 THEN 0.6 WHEN 2 THEN 0.3 ELSE 0.1 END AS w
				FROM recent
			) AS t
		), lifetime AS (
			SELECT count(*) AS cnt, max(created_at) AS last_review
			FROM canteen_dish_review_events
			WHERE dish_id = ? AND user_id = ? AND status = ?
		)
		INSERT INTO canteen_dish_rating_summaries
			(dish_id, user_id, weighted_score, window_size, lifetime_event_count, last_reviewed_at, computed_at)
		SELECT ?, ?, round(weighted.score::numeric, 2), weighted.window_size,
		       lifetime.cnt, lifetime.last_review, now()
		FROM weighted, lifetime
		ON CONFLICT (dish_id, user_id) DO UPDATE SET
			weighted_score = EXCLUDED.weighted_score,
			window_size = EXCLUDED.window_size,
			lifetime_event_count = EXCLUDED.lifetime_event_count,
			last_reviewed_at = EXCLUDED.last_reviewed_at,
			computed_at = now()`,
		dishID, userID, string(canteen.ReviewStatusActive),
		dishID, userID, string(canteen.ReviewStatusActive),
		dishID, userID,
	).Error
}

func recomputeCanteenDishStats(tx *gorm.DB, dishID int, priorCount int) error {
	if priorCount <= 0 {
		priorCount = canteen.DefaultPriorCount
	}
	return tx.Exec(`
		WITH summaries AS (
			SELECT count(*)::int AS cnt, COALESCE(avg(weighted_score), 0) AS avg
			FROM canteen_dish_rating_summaries
			WHERE dish_id = ?
		), events AS (
			SELECT count(*)::int AS events,
			       COALESCE(avg(taste), 0) AS taste,
			       COALESCE(avg(portion), 0) AS portion,
			       COALESCE(avg(value), 0) AS value
			FROM canteen_dish_review_events
			WHERE dish_id = ? AND status = ?
		), global_avg AS (
			SELECT COALESCE(avg(weighted_score), 0) AS value
			FROM canteen_dish_rating_summaries
		)
		UPDATE canteen_dishes AS d SET
			rating_count = summaries.cnt,
			rating_avg = round(summaries.avg::numeric, 2),
			rating_score = CASE
				WHEN summaries.cnt = 0 THEN 0
				ELSE round(((summaries.avg * summaries.cnt + global_avg.value * ?) / (summaries.cnt + ?))::numeric, 2)
			END,
			taste_avg = round(events.taste::numeric, 2),
			portion_avg = round(events.portion::numeric, 2),
			value_avg = round(events.value::numeric, 2),
			review_event_count = events.events,
			rating_updated_at = now(),
			updated_at = now()
		FROM summaries, events, global_avg
		WHERE d.id = ?`,
		dishID, dishID, string(canteen.ReviewStatusActive),
		priorCount, priorCount, dishID,
	).Error
}
