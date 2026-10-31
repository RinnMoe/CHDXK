package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"jcourse/internal/domain/canteen"
)

// CanteenVoteRepository stores one helpfulness vote per user and review.
type CanteenVoteRepository struct {
	db *gorm.DB
}

func NewCanteenVoteRepository(db *gorm.DB) *CanteenVoteRepository {
	return &CanteenVoteRepository{db: db}
}

func (r *CanteenVoteRepository) Get(ctx context.Context, reviewEventID, userID int) (*canteen.Vote, error) {
	var entity CanteenReviewVoteEntity
	err := r.db.WithContext(ctx).
		Where("review_event_id = ? AND user_id = ?", reviewEventID, userID).
		Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &canteen.Vote{
		ReviewEventID: entity.ReviewEventID,
		UserID:        entity.UserID,
		VoteType:      canteen.VoteType(entity.VoteType),
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}, nil
}

func (r *CanteenVoteRepository) Save(ctx context.Context, vote *canteen.Vote) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			INSERT INTO canteen_review_votes (review_event_id, user_id, vote_type, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (review_event_id, user_id) DO UPDATE SET
				vote_type = EXCLUDED.vote_type,
				updated_at = EXCLUDED.updated_at`,
			vote.ReviewEventID, vote.UserID, int(vote.VoteType), vote.CreatedAt, vote.UpdatedAt,
		).Error; err != nil {
			return err
		}
		return recomputeCanteenVoteCounters(tx, vote.ReviewEventID)
	})
}

func (r *CanteenVoteRepository) Delete(ctx context.Context, reviewEventID, userID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DELETE FROM canteen_review_votes WHERE review_event_id = ? AND user_id = ?",
			reviewEventID, userID,
		).Error; err != nil {
			return err
		}
		return recomputeCanteenVoteCounters(tx, reviewEventID)
	})
}

func recomputeCanteenVoteCounters(tx *gorm.DB, reviewEventID int) error {
	return tx.Exec(`
		UPDATE canteen_review_events AS e SET
			like_count = (SELECT count(*) FROM canteen_review_votes v WHERE v.review_event_id = e.id AND v.vote_type = 1),
			dislike_count = (SELECT count(*) FROM canteen_review_votes v WHERE v.review_event_id = e.id AND v.vote_type = -1)
		WHERE e.id = ?`, reviewEventID).Error
}
