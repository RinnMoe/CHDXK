package application

import (
	"context"
	"errors"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
	"jcourse/internal/domain/canteen/policy"
	"jcourse/pkg/apperr"
)

// CreateCanteenReviewCommand is one visit review submission.
type CreateCanteenReviewCommand struct {
	StoreID         int
	VisitedAt       time.Time
	Taste           int
	Value           int
	Queue           int
	Hygiene         int
	Service         int
	Content         string
	Tags            []string
	ClientRequestID string
	MediaIDs        []int
}

// UpdateCanteenReviewCommand edits the editable subset of a visit.
type UpdateCanteenReviewCommand struct {
	ReviewID          int
	VisitedAt         time.Time
	Taste             int
	Value             int
	Queue             int
	Hygiene           int
	Service           int
	Content           string
	Tags              []string
	ExpectedUpdatedAt time.Time
}

// CanteenReviewCommandService handles visit review writes. Aggregates are
// maintained by the repository inside the same transaction.
type CanteenReviewCommandService struct {
	stores          canteen.StoreRepository
	reviews         canteen.ReviewRepository
	votes           canteen.VoteRepository
	reviewQuery     canteen.ReviewQuery
	settings        CanteenSettingsProvider
	frequency       policy.Config
	maxDailyReviews int
	media           *CanteenMediaCommandService
	now             func() time.Time
}

func NewCanteenReviewCommandService(
	stores canteen.StoreRepository,
	reviews canteen.ReviewRepository,
	votes canteen.VoteRepository,
	reviewQuery canteen.ReviewQuery,
	settings CanteenSettingsProvider,
	frequency policy.Config,
	maxDailyReviews int,
	media *CanteenMediaCommandService,
) *CanteenReviewCommandService {
	if maxDailyReviews <= 0 {
		maxDailyReviews = 5
	}
	return &CanteenReviewCommandService{
		stores:          stores,
		reviews:         reviews,
		votes:           votes,
		reviewQuery:     reviewQuery,
		settings:        settings,
		frequency:       frequency,
		maxDailyReviews: maxDailyReviews,
		media:           media,
		now:             time.Now,
	}
}

func (s *CanteenReviewCommandService) runtime(ctx context.Context) (CanteenRuntimeConfig, error) {
	if s.settings == nil {
		return CanteenRuntimeConfig{Enabled: true, ReviewsEnabled: true}, nil
	}
	return s.settings.CanteenRuntimeConfig(ctx)
}

func (s *CanteenReviewCommandService) requireWritable(ctx context.Context) (CanteenRuntimeConfig, error) {
	config, err := s.runtime(ctx)
	if err != nil {
		return CanteenRuntimeConfig{}, err
	}
	if !config.Enabled {
		return CanteenRuntimeConfig{}, apperr.NotFound("食堂评价暂未开放")
	}
	if !config.ReviewsEnabled {
		return CanteenRuntimeConfig{}, apperr.Forbidden("到店评价暂未开放")
	}
	return config, nil
}

func (s *CanteenReviewCommandService) Create(
	ctx context.Context,
	user *auth.User,
	cmd CreateCanteenReviewCommand,
) (*canteen.ReviewEvent, bool, error) {
	if _, err := s.requireWritable(ctx); err != nil {
		return nil, false, err
	}
	if !canteen.CanCreateReview(user) {
		if user != nil && user.IsSuspended() {
			return nil, false, apperr.ErrUserSuspended
		}
		return nil, false, auth.ErrMuted
	}
	if cmd.ClientRequestID != "" {
		existing, err := s.reviews.FindByClientRequestID(ctx, user.ID, cmd.ClientRequestID)
		if err != nil {
			return nil, false, err
		}
		if existing != nil {
			return existing, false, nil
		}
	}
	store, err := s.stores.Get(ctx, cmd.StoreID)
	if err != nil {
		return nil, false, err
	}
	if store == nil {
		return nil, false, canteen.ErrStoreNotFound
	}
	if !store.CanAcceptReview() {
		return nil, false, canteen.ErrStoreNotActive
	}
	now := s.now().Truncate(time.Microsecond)
	if err := validateCanteenVisitTime(cmd.VisitedAt, now); err != nil {
		return nil, false, err
	}
	count, err := s.reviewQuery.CountSince(ctx, user.ID, now.Add(-24*time.Hour))
	if err != nil {
		return nil, false, err
	}
	if count >= s.maxDailyReviews {
		return nil, false, canteen.ErrDailyReviewLimit
	}
	policyChecker := policy.NewFrequencyPolicy(s.reviewQuery, s.frequency)
	if err := policyChecker.Check(ctx, user, cmd.Content); err != nil {
		return nil, false, err
	}

	review := &canteen.ReviewEvent{
		StoreID:   cmd.StoreID,
		UserID:    user.ID,
		VisitedAt: cmd.VisitedAt,
		Dimensions: canteen.Dimensions{
			Taste:   cmd.Taste,
			Value:   cmd.Value,
			Queue:   cmd.Queue,
			Hygiene: cmd.Hygiene,
			Service: cmd.Service,
		},
		Content:         cmd.Content,
		Tags:            cmd.Tags,
		ClientRequestID: cmd.ClientRequestID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if review.VisitedAt.IsZero() {
		review.VisitedAt = now
	}
	review.ApplyCreate(canteen.DefaultStoreWeights)
	if err := review.Validate(); err != nil {
		return nil, false, err
	}
	if s.media != nil && len(cmd.MediaIDs) > 0 {
		if _, err := s.media.ValidateAttachments(ctx, user, cmd.MediaIDs); err != nil {
			return nil, false, err
		}
	}
	if err := s.reviews.Create(ctx, review); err != nil {
		if errors.Is(err, canteen.ErrReviewConflict) && cmd.ClientRequestID != "" {
			existing, findErr := s.reviews.FindByClientRequestID(ctx, user.ID, cmd.ClientRequestID)
			if findErr != nil {
				return nil, false, findErr
			}
			if existing != nil {
				return existing, false, nil
			}
		}
		return nil, false, err
	}
	if s.media != nil && len(cmd.MediaIDs) > 0 {
		if err := s.media.Attach(ctx, user, review.ID, cmd.MediaIDs); err != nil {
			_ = s.reviews.SoftDelete(ctx, review.ID, now)
			return nil, false, err
		}
	}
	return review, true, nil
}

func (s *CanteenReviewCommandService) Update(
	ctx context.Context,
	user *auth.User,
	cmd UpdateCanteenReviewCommand,
) (*canteen.ReviewEvent, error) {
	if _, err := s.requireWritable(ctx); err != nil {
		return nil, err
	}
	review, err := s.reviews.Get(ctx, cmd.ReviewID)
	if err != nil {
		return nil, err
	}
	if review == nil {
		return nil, canteen.ErrReviewNotFound
	}
	guardian := canteen.NewGuardian(user, review)
	if !guardian.CanUpdate() {
		return nil, canteen.ErrCannotEditReview
	}
	if !cmd.ExpectedUpdatedAt.IsZero() && !review.UpdatedAt.Equal(cmd.ExpectedUpdatedAt) {
		return nil, canteen.ErrReviewConflict
	}
	if cmd.ExpectedUpdatedAt.IsZero() {
		return nil, apperr.BadRequest("缺少评价版本，请刷新后重试")
	}
	now := s.now().Truncate(time.Microsecond)
	if !now.After(review.UpdatedAt) {
		now = review.UpdatedAt.Add(time.Microsecond)
	}
	if err := validateCanteenVisitTime(cmd.VisitedAt, now); err != nil {
		return nil, err
	}
	expectedUpdatedAt := review.UpdatedAt
	revision := review.MakeRevision(user.ID, now)
	update := canteen.UpdateReview{
		VisitedAt: cmd.VisitedAt,
		Dimensions: canteen.Dimensions{
			Taste:   cmd.Taste,
			Value:   cmd.Value,
			Queue:   cmd.Queue,
			Hygiene: cmd.Hygiene,
			Service: cmd.Service,
		},
		Content: cmd.Content,
		Tags:    cmd.Tags,
		Now:     now,
	}
	if err := review.ApplyUpdate(update, canteen.DefaultStoreWeights); err != nil {
		return nil, err
	}
	if err := s.reviews.Update(ctx, review, revision, expectedUpdatedAt); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *CanteenReviewCommandService) Delete(ctx context.Context, user *auth.User, reviewID int) error {
	if _, err := s.requireWritable(ctx); err != nil {
		return err
	}
	review, err := s.reviews.Get(ctx, reviewID)
	if err != nil {
		return err
	}
	if review == nil {
		return canteen.ErrReviewNotFound
	}
	if !canteen.NewGuardian(user, review).CanDelete() {
		return canteen.ErrCannotModerate
	}
	if review.Status == canteen.ReviewStatusDeleted {
		return nil
	}
	return s.reviews.SoftDelete(ctx, reviewID, s.now())
}

func (s *CanteenReviewCommandService) Hide(
	ctx context.Context,
	user *auth.User,
	reviewID int,
	reason string,
) error {
	review, err := s.reviews.Get(ctx, reviewID)
	if err != nil {
		return err
	}
	if review == nil {
		return canteen.ErrReviewNotFound
	}
	if !canteen.NewGuardian(user, review).CanHide() {
		return canteen.ErrCannotModerate
	}
	return s.reviews.Hide(ctx, reviewID, user.ID, reason, s.now())
}

// Vote upserts or clears the caller's helpfulness vote.
func (s *CanteenReviewCommandService) Vote(
	ctx context.Context,
	user *auth.User,
	reviewID int,
	voteType int,
) error {
	if _, err := s.requireWritable(ctx); err != nil {
		return err
	}
	review, err := s.reviews.Get(ctx, reviewID)
	if err != nil {
		return err
	}
	if review == nil || review.Status != canteen.ReviewStatusActive {
		return canteen.ErrReviewNotFound
	}
	if review.UserID == user.ID {
		return canteen.ErrVoteOwnReview
	}
	now := s.now()
	if voteType == 0 {
		return s.votes.Delete(ctx, reviewID, user.ID)
	}
	vote := canteen.VoteType(voteType)
	if !vote.Valid() {
		return canteen.ErrInvalidVoteType
	}
	return s.votes.Save(ctx, &canteen.Vote{
		ReviewEventID: reviewID,
		UserID:        user.ID,
		VoteType:      vote,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
}

// RefreshRankings recomputes every store from persisted summaries.
func (s *CanteenReviewCommandService) RefreshRankings(ctx context.Context) error {
	return s.reviews.RefreshAllStoreStats(ctx)
}

// validateCanteenVisitTime keeps client-supplied visit times inside a
// defensible window. Future visits would poison the recent-visit weighting;
// arbitrarily old backfill would rewrite historical rankings.
func validateCanteenVisitTime(visitedAt, now time.Time) error {
	if visitedAt.IsZero() {
		return nil
	}
	if visitedAt.After(now.Add(24 * time.Hour)) {
		return apperr.BadRequest("到店时间不能晚于当前时间")
	}
	if visitedAt.Before(now.AddDate(-1, 0, 0)) {
		return apperr.BadRequest("到店时间不能早于一年前")
	}
	return nil
}
