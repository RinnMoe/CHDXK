package review

import (
	"context"
	"errors"
	"time"
)

const (
	VoteLike    = 1
	VoteDislike = -1
)

type VoteConfig struct {
	MaxDailyVotes int
}

func DefaultVoteConfig() VoteConfig {
	return VoteConfig{MaxDailyVotes: 50}
}

type Vote struct {
	ReviewID  int
	UserID    int
	VoteType  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type VoteResult struct {
	CourseID int
	Changed  bool
}

var ErrDailyVoteLimitReached = errors.New("daily vote limit reached")

var ErrInvalidVoteType = errors.New("invalid vote type")

type VoteService struct {
	reviewRepo ReviewRepository
	voteRepo   VoteRepository
	config     VoteConfig
}

func NewVoteService(reviewRepo ReviewRepository, voteRepo VoteRepository, config VoteConfig) *VoteService {
	if config.MaxDailyVotes <= 0 {
		config.MaxDailyVotes = DefaultVoteConfig().MaxDailyVotes
	}
	return &VoteService{reviewRepo: reviewRepo, voteRepo: voteRepo, config: config}
}

func (s *VoteService) Vote(ctx context.Context, userID, reviewID, voteType int, now time.Time) (*VoteResult, error) {
	if voteType != 0 && voteType != VoteLike && voteType != VoteDislike {
		return nil, ErrInvalidVoteType
	}
	reviewView, err := s.reviewRepo.Get(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if reviewView == nil {
		return nil, ErrReviewNotFound
	}

	todayCount, err := s.voteRepo.CountTodayByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if todayCount >= int64(s.config.MaxDailyVotes) {
		return nil, ErrDailyVoteLimitReached
	}

	existing, err := s.voteRepo.FindByReviewAndUser(ctx, reviewID, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.VoteType == voteType {
		return &VoteResult{CourseID: reviewView.CourseID, Changed: false}, nil
	}
	if voteType == 0 {
		if existing == nil {
			return &VoteResult{CourseID: reviewView.CourseID, Changed: false}, nil
		}
		if err := s.voteRepo.Delete(ctx, reviewID, userID); err != nil {
			return nil, err
		}
		return &VoteResult{CourseID: reviewView.CourseID, Changed: true}, nil
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
	if err := s.voteRepo.Save(ctx, v); err != nil {
		return nil, err
	}
	return &VoteResult{CourseID: reviewView.CourseID, Changed: true}, nil
}

type VoteRepository interface {
	FindByReviewAndUser(ctx context.Context, reviewID, userID int) (*Vote, error)
	FindByReviewsAndUser(ctx context.Context, reviewIDs []int, userID int) (map[int]Vote, error)
	CountTodayByUser(ctx context.Context, userID int) (int64, error)
	Save(ctx context.Context, v *Vote) error
	Delete(ctx context.Context, reviewID, userID int) error
}
