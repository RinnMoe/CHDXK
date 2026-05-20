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

type VoteRepository interface {
	FindByReviewAndUser(ctx context.Context, reviewID, userID int) (*Vote, error)
	CountTodayByUser(ctx context.Context, userID int) (int64, error)
	Save(ctx context.Context, v *Vote) error
	Delete(ctx context.Context, reviewID, userID int) error
}
