package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"jcourse/internal/domain/review"
	"jcourse/pkg/logx"
)

type ReviewVoteRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewReviewVoteRepository(db *gorm.DB, cache ...*redis.Client) *ReviewVoteRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &ReviewVoteRepository{db: db, cache: client}
}

func newVoteEntity(v *review.Vote) ReviewVoteEntity {
	return ReviewVoteEntity{
		ReviewID:  v.ReviewID,
		UserID:    v.UserID,
		VoteType:  v.VoteType,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

func newVoteDomain(e *ReviewVoteEntity) review.Vote {
	return review.Vote{
		ReviewID:  e.ReviewID,
		UserID:    e.UserID,
		VoteType:  e.VoteType,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func (r *ReviewVoteRepository) FindByReviewAndUser(ctx context.Context, reviewID, userID int) (*review.Vote, error) {
	key := cacheKey("review_vote", reviewID, userID)
	if cached, ok := cacheGetJSON[review.Vote](ctx, r.cache, key); ok {
		return cached, nil
	}

	e, err := gorm.G[ReviewVoteEntity](r.db).Where("review_id = ? AND user_id = ?", reviewID, userID).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	v := newVoteDomain(&e)
	cacheSetJSON(ctx, r.cache, key, &v)
	return &v, nil
}

func (r *ReviewVoteRepository) FindByReviewsAndUser(ctx context.Context, reviewIDs []int, userID int) (map[int]review.Vote, error) {
	votes := make(map[int]review.Vote)
	if userID == 0 || len(reviewIDs) == 0 {
		return votes, nil
	}

	ids := make([]int, 0, len(reviewIDs))
	seen := make(map[int]struct{}, len(reviewIDs))
	for _, id := range reviewIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return votes, nil
	}

	missing := ids
	if r.cache != nil {
		keys := make([]string, len(ids))
		for i, reviewID := range ids {
			keys[i] = cacheKey("review_vote", reviewID, userID)
		}
		if cached, err := r.cache.MGet(ctx, keys...).Result(); err == nil {
			missing = make([]int, 0, len(ids))
			for i, value := range cached {
				if value == nil {
					missing = append(missing, ids[i])
					continue
				}

				data, ok := value.(string)
				if !ok {
					cacheDelete(ctx, r.cache, keys[i])
					missing = append(missing, ids[i])
					continue
				}

				var vote review.Vote
				if err := json.Unmarshal([]byte(data), &vote); err != nil {
					cacheDelete(ctx, r.cache, keys[i])
					missing = append(missing, ids[i])
					continue
				}
				votes[vote.ReviewID] = vote
			}
		} else {
			logx.Warn(ctx, "cache access failed", "operation", "mget", "keys", keys, "err", err)
		}
	}

	if len(missing) == 0 {
		return votes, nil
	}

	var entities []ReviewVoteEntity
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND review_id IN ?", userID, missing).
		Find(&entities).Error; err != nil {
		return nil, err
	}

	for _, entity := range entities {
		vote := newVoteDomain(&entity)
		votes[vote.ReviewID] = vote
		cacheSetJSON(ctx, r.cache, cacheKey("review_vote", vote.ReviewID, userID), &vote)
	}

	return votes, nil
}

func (r *ReviewVoteRepository) CountTodayByUser(ctx context.Context, userID int) (int64, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var count int64
	err := r.db.Model(&ReviewVoteEntity{}).Where("user_id = ? AND updated_at >= ?", userID, startOfDay).Count(&count).Error
	return count, err
}

func (r *ReviewVoteRepository) Save(ctx context.Context, v *review.Vote) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		e := newVoteEntity(v)
		if err := tx.Save(&e).Error; err != nil {
			return err
		}
		return r.updateReviewVoteCounts(tx, v.ReviewID)
	}); err != nil {
		return err
	}
	cacheDelete(ctx, r.cache,
		cacheKey("review", v.ReviewID, "view"),
		cacheKey("review_vote", v.ReviewID, v.UserID),
	)
	return nil
}

func (r *ReviewVoteRepository) Delete(ctx context.Context, reviewID, userID int) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("review_id = ? AND user_id = ?", reviewID, userID).Delete(&ReviewVoteEntity{}).Error; err != nil {
			return err
		}
		return r.updateReviewVoteCounts(tx, reviewID)
	}); err != nil {
		return err
	}
	cacheDelete(ctx, r.cache,
		cacheKey("review", reviewID, "view"),
		cacheKey("review_vote", reviewID, userID),
	)
	return nil
}

func (r *ReviewVoteRepository) updateReviewVoteCounts(tx *gorm.DB, reviewID int) error {
	return tx.Exec(`UPDATE reviews SET
		like_count = (SELECT COUNT(*) FROM review_votes WHERE review_id = ? AND vote_type = 1),
		dislike_count = (SELECT COUNT(*) FROM review_votes WHERE review_id = ? AND vote_type = -1)
		WHERE id = ?`, reviewID, reviewID, reviewID).Error
}

var _ review.VoteRepository = (*ReviewVoteRepository)(nil)
