package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/review"
)

type ReviewVoteRepository struct {
	db *gorm.DB
}

func NewReviewVoteRepository(db *gorm.DB) *ReviewVoteRepository {
	return &ReviewVoteRepository{db: db}
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
	e, err := gorm.G[ReviewVoteEntity](r.db).Where("review_id = ? AND user_id = ?", reviewID, userID).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	v := newVoteDomain(&e)
	return &v, nil
}

func (r *ReviewVoteRepository) CountTodayByUser(ctx context.Context, userID int) (int64, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var count int64
	err := r.db.Model(&ReviewVoteEntity{}).Where("user_id = ? AND updated_at >= ?", userID, startOfDay).Count(&count).Error
	return count, err
}

func (r *ReviewVoteRepository) Save(ctx context.Context, v *review.Vote) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		e := newVoteEntity(v)
		if err := tx.Save(&e).Error; err != nil {
			return err
		}
		return r.updateReviewVoteCounts(tx, v.ReviewID)
	})
}

func (r *ReviewVoteRepository) Delete(ctx context.Context, reviewID, userID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("review_id = ? AND user_id = ?", reviewID, userID).Delete(&ReviewVoteEntity{}).Error; err != nil {
			return err
		}
		return r.updateReviewVoteCounts(tx, reviewID)
	})
}

func (r *ReviewVoteRepository) updateReviewVoteCounts(tx *gorm.DB, reviewID int) error {
	return tx.Exec(`UPDATE reviews SET
		like_count = (SELECT COUNT(*) FROM review_votes WHERE review_id = ? AND vote_type = 1),
		dislike_count = (SELECT COUNT(*) FROM review_votes WHERE review_id = ? AND vote_type = -1)
		WHERE id = ?`, reviewID, reviewID, reviewID).Error
}

var _ review.VoteRepository = (*ReviewVoteRepository)(nil)
