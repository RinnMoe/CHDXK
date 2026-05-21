package review

import (
	"context"
	"errors"
	"time"
)

const (
	VoteLike    = 1
	VoteDislike = -1

	MaxDailyVotes = 50
)

type Vote struct {
	ReviewID  int
	UserID    int
	VoteType  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

var ErrDailyVoteLimitReached = errors.New("daily vote limit reached")

var ErrInvalidVoteType = errors.New("invalid vote type")

type VoteService struct {
	reviewRepo ReviewRepository
	voteRepo   VoteRepository
}

func NewVoteService(reviewRepo ReviewRepository, voteRepo VoteRepository) *VoteService {
	return &VoteService{reviewRepo: reviewRepo, voteRepo: voteRepo}
}

func (s *VoteService) Vote(ctx context.Context, userID, reviewID, voteType int, now time.Time) error {
	if voteType != 0 && voteType != VoteLike && voteType != VoteDislike {
		return ErrInvalidVoteType
	}
	reviewView, err := s.reviewRepo.Get(ctx, reviewID)
	if err != nil {
		return err
	}
	if reviewView == nil {
		return ErrReviewNotFound
	}

	todayCount, err := s.voteRepo.CountTodayByUser(ctx, userID)
	if err != nil {
		return err
	}
	if todayCount >= MaxDailyVotes {
		return ErrDailyVoteLimitReached
	}

	existing, err := s.voteRepo.FindByReviewAndUser(ctx, reviewID, userID)
	if err != nil {
		return err
	}
	if existing != nil && existing.VoteType == voteType {
		return nil
	}
	if voteType == 0 {
		if existing == nil {
			return nil
		}
		return s.voteRepo.Delete(ctx, reviewID, userID)
	}
	v := &Vote{
		ReviewID:  reviewID,
		UserID:    userID,
		VoteType:  voteType,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if existing != nil {
		v.CreatedAt = existing.CreatedAt
	}
	return s.voteRepo.Save(ctx, v)
}

type VoteRepository interface {
	FindByReviewAndUser(ctx context.Context, reviewID, userID int) (*Vote, error)
	CountTodayByUser(ctx context.Context, userID int) (int64, error)
	Save(ctx context.Context, v *Vote) error
	Delete(ctx context.Context, reviewID, userID int) error
}
