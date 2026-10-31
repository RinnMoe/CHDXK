package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"jcourse/internal/domain/canteen"
)

// CanteenReviewRepository persists visit events and keeps the per-user and
// per-store aggregates consistent in the same transaction.
type CanteenReviewRepository struct {
	db         *gorm.DB
	cache      *redis.Client
	priorCount int
}

func NewCanteenReviewRepository(db *gorm.DB, cache ...*redis.Client) *CanteenReviewRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &CanteenReviewRepository{
		db:         db,
		cache:      client,
		priorCount: canteen.DefaultPriorCount,
	}
}

func (r *CanteenReviewRepository) SetPriorCount(prior int) {
	if prior > 0 {
		r.priorCount = prior
	}
}

func (r *CanteenReviewRepository) Create(ctx context.Context, review *canteen.ReviewEvent) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockCanteenStore(tx, review.StoreID); err != nil {
			return err
		}
		entity := newCanteenReviewEntity(review)
		if err := tx.Create(&entity).Error; err != nil {
			if isCanteenUniqueViolation(err) {
				return canteen.ErrReviewConflict
			}
			return err
		}
		review.ID = entity.ID
		review.CreatedAt = entity.CreatedAt
		review.UpdatedAt = entity.UpdatedAt
		if err := recomputeCanteenUserSummary(tx, review.StoreID, review.UserID); err != nil {
			return err
		}
		return recomputeCanteenStoreStats(tx, review.StoreID, r.priorCount)
	})
	if err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenReviewRepository) Update(ctx context.Context, review *canteen.ReviewEvent, revision canteen.Revision) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockCanteenStore(tx, review.StoreID); err != nil {
			return err
		}
		revisionEntity := newCanteenRevisionEntity(&revision)
		if err := tx.Create(&revisionEntity).Error; err != nil {
			return err
		}
		result := tx.Model(&CanteenReviewEventEntity{}).Where("id = ?", review.ID).Updates(map[string]any{
			"visited_at":            review.VisitedAt,
			"taste":                 review.Dimensions.Taste,
			"value":                 review.Dimensions.Value,
			"queue":                 review.Dimensions.Queue,
			"hygiene":               review.Dimensions.Hygiene,
			"service":               review.Dimensions.Service,
			"overall_score":         review.OverallScore,
			"score_formula_version": review.ScoreFormulaVersion,
			"content":               review.Content,
			"tags":                  canteenStringArray(review.Tags),
			"updated_at":            review.UpdatedAt,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return canteen.ErrReviewNotFound
		}
		if err := recomputeCanteenUserSummary(tx, review.StoreID, review.UserID); err != nil {
			return err
		}
		return recomputeCanteenStoreStats(tx, review.StoreID, r.priorCount)
	})
	if err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenReviewRepository) Get(ctx context.Context, id int) (*canteen.ReviewEvent, error) {
	var entity CanteenReviewEventEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return newCanteenReviewDomain(&entity), nil
}

func (r *CanteenReviewRepository) FindByClientRequestID(ctx context.Context, userID int, requestID string) (*canteen.ReviewEvent, error) {
	if requestID == "" {
		return nil, nil
	}
	var entity CanteenReviewEventEntity
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND client_request_id = ?", userID, requestID).
		Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return newCanteenReviewDomain(&entity), nil
}

func (r *CanteenReviewRepository) SoftDelete(ctx context.Context, id int, now time.Time) error {
	return r.changeStatus(ctx, id, func(tx *gorm.DB) error {
		result := tx.Model(&CanteenReviewEventEntity{}).Where("id = ?", id).Updates(map[string]any{
			"status":     string(canteen.ReviewStatusDeleted),
			"updated_at": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return canteen.ErrReviewNotFound
		}
		return nil
	})
}

func (r *CanteenReviewRepository) Hide(ctx context.Context, id int, byUserID int, reason string, now time.Time) error {
	return r.changeStatus(ctx, id, func(tx *gorm.DB) error {
		result := tx.Model(&CanteenReviewEventEntity{}).Where("id = ?", id).Updates(map[string]any{
			"status":        string(canteen.ReviewStatusHidden),
			"hidden_by":     byUserID,
			"hidden_reason": reason,
			"hidden_at":     now,
			"updated_at":    now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return canteen.ErrReviewNotFound
		}
		return nil
	})
}

// changeStatus applies a status mutation and recomputes the joined store.
func (r *CanteenReviewRepository) changeStatus(ctx context.Context, id int, update func(tx *gorm.DB) error) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var storeID int
		if err := tx.Raw("SELECT store_id FROM canteen_review_events WHERE id = ?", id).Scan(&storeID).Error; err != nil {
			return err
		}
		if storeID == 0 {
			return canteen.ErrReviewNotFound
		}
		if err := lockCanteenStore(tx, storeID); err != nil {
			return err
		}
		if err := update(tx); err != nil {
			return err
		}
		var userID int
		if err := tx.Raw("SELECT user_id FROM canteen_review_events WHERE id = ?", id).Scan(&userID).Error; err != nil {
			return err
		}
		if err := recomputeCanteenUserSummary(tx, storeID, userID); err != nil {
			return err
		}
		return recomputeCanteenStoreStats(tx, storeID, r.priorCount)
	})
	if err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenReviewRepository) RefreshStoreStats(ctx context.Context, storeID int) error {
	if err := recomputeCanteenStoreStats(r.db.WithContext(ctx), storeID, r.priorCount); err != nil {
		return err
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenReviewRepository) RefreshAllStoreStats(ctx context.Context) error {
	var storeIDs []int
	if err := r.db.WithContext(ctx).Model(&CanteenStoreEntity{}).Order("id").Pluck("id", &storeIDs).Error; err != nil {
		return err
	}
	for _, storeID := range storeIDs {
		if err := recomputeCanteenStoreStats(r.db.WithContext(ctx), storeID, r.priorCount); err != nil {
			return err
		}
	}
	r.invalidate(ctx)
	return nil
}

func (r *CanteenReviewRepository) invalidate(ctx context.Context) {
	cacheDeletePattern(ctx, r.cache, cacheKey("canteen", "*"))
}

func lockCanteenStore(tx *gorm.DB, storeID int) error {
	var id int
	if err := tx.Raw("SELECT id FROM canteen_stores WHERE id = ? FOR UPDATE", storeID).Scan(&id).Error; err != nil {
		return err
	}
	if id == 0 {
		return canteen.ErrStoreNotFound
	}
	return nil
}

// recomputeCanteenUserSummary keeps the latest three active events, weighted
// 0.6/0.3/0.1 and renormalized when fewer events exist.
func recomputeCanteenUserSummary(tx *gorm.DB, storeID, userID int) error {
	var activeCount int
	if err := tx.Raw(`
		SELECT count(*) FROM canteen_review_events
		WHERE store_id = ? AND user_id = ? AND status = ?`,
		storeID, userID, string(canteen.ReviewStatusActive)).Scan(&activeCount).Error; err != nil {
		return err
	}
	if activeCount == 0 {
		return tx.Exec(
			"DELETE FROM canteen_user_rating_summaries WHERE store_id = ? AND user_id = ?",
			storeID, userID,
		).Error
	}
	return tx.Exec(`
		WITH recent AS (
			SELECT overall_score, visited_at, id,
			       row_number() OVER (ORDER BY visited_at DESC, id DESC) AS rn
			FROM canteen_review_events
			WHERE store_id = ? AND user_id = ? AND status = ?
			ORDER BY visited_at DESC, id DESC
			LIMIT 3
		), weighted AS (
			SELECT sum(overall_score * w) / sum(w) AS score, count(*) AS window_size
			FROM (
				SELECT overall_score,
				       CASE rn WHEN 1 THEN 0.6 WHEN 2 THEN 0.3 ELSE 0.1 END AS w
				FROM recent
			) AS t
		), lifetime AS (
			SELECT count(*) AS cnt, max(visited_at) AS last_visit
			FROM canteen_review_events
			WHERE store_id = ? AND user_id = ? AND status = ?
		)
		INSERT INTO canteen_user_rating_summaries
			(store_id, user_id, weighted_score, window_size, lifetime_event_count, last_visited_at, computed_at)
		SELECT ?, ?, round(weighted.score::numeric, 2), weighted.window_size,
		       lifetime.cnt, lifetime.last_visit, now()
		FROM weighted, lifetime
		ON CONFLICT (store_id, user_id) DO UPDATE SET
			weighted_score = EXCLUDED.weighted_score,
			window_size = EXCLUDED.window_size,
			lifetime_event_count = EXCLUDED.lifetime_event_count,
			last_visited_at = EXCLUDED.last_visited_at,
			computed_at = now()`,
		storeID, userID, string(canteen.ReviewStatusActive),
		storeID, userID, string(canteen.ReviewStatusActive),
		storeID, userID,
	).Error
}

// recomputeCanteenStoreStats recomputes one store from its summaries and
// active events. Bayesian shrinkage uses the platform-wide summary mean.
func recomputeCanteenStoreStats(tx *gorm.DB, storeID int, priorCount int) error {
	if priorCount <= 0 {
		priorCount = canteen.DefaultPriorCount
	}
	return tx.Exec(`
		WITH summaries AS (
			SELECT count(*)::int AS cnt, COALESCE(avg(weighted_score), 0) AS avg
			FROM canteen_user_rating_summaries
			WHERE store_id = ?
		), events AS (
			SELECT count(*)::int AS events,
			       COALESCE(avg(taste), 0) AS taste,
			       COALESCE(avg(value), 0) AS value,
			       COALESCE(avg(queue), 0) AS queue,
			       COALESCE(avg(hygiene), 0) AS hygiene,
			       COALESCE(avg(service), 0) AS service,
			       max(visited_at) AS last_visit
			FROM canteen_review_events
			WHERE store_id = ? AND status = ?
		), global_avg AS (
			SELECT COALESCE(avg(weighted_score), 0) AS value
			FROM canteen_user_rating_summaries
		)
		UPDATE canteen_stores AS s SET
			rating_count = summaries.cnt,
			rating_avg = round(summaries.avg::numeric, 2),
			rating_score = CASE
				WHEN summaries.cnt = 0 THEN 0
				ELSE round(((summaries.avg * summaries.cnt + global_avg.value * ?) / (summaries.cnt + ?))::numeric, 2)
			END,
			taste_avg = round(events.taste::numeric, 2),
			value_avg = round(events.value::numeric, 2),
			queue_avg = round(events.queue::numeric, 2),
			hygiene_avg = round(events.hygiene::numeric, 2),
			service_avg = round(events.service::numeric, 2),
			review_event_count = events.events,
			last_review_at = events.last_visit,
			rating_updated_at = now(),
			updated_at = now()
		FROM summaries, events, global_avg
		WHERE s.id = ?`,
		storeID, storeID, string(canteen.ReviewStatusActive),
		priorCount, priorCount, storeID,
	).Error
}
