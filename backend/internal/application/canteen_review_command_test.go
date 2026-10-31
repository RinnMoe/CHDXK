package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
	"jcourse/internal/domain/canteen/policy"
)

func newCanteenReviewCommandFixture(t *testing.T) (*CanteenReviewCommandService, *canteen.MockReviewRepository, *canteen.MockStoreRepository, *canteen.MockVoteRepository, *canteen.MockReviewQuery) {
	t.Helper()
	stores := canteen.NewMockStoreRepository()
	store := &canteen.Store{HallID: 1, Name: "老王面馆", Status: canteen.StoreStatusActive}
	store.ApplyDefaults()
	if err := stores.Create(context.Background(), store); err != nil {
		t.Fatalf("seed store: %v", err)
	}
	reviews := canteen.NewMockReviewRepository()
	votes := canteen.NewMockVoteRepository()
	query := &canteen.MockReviewQuery{}
	service := NewCanteenReviewCommandService(
		stores,
		reviews,
		votes,
		query,
		StaticCanteenSettingsProvider{Runtime: CanteenRuntimeConfig{Enabled: true, ReviewsEnabled: true}},
		policy.Config{Window: time.Hour, MaxReviews: 10, SimilarityRatio: 0.85},
		5,
		nil,
	)
	return service, reviews, stores, votes, query
}

func validCanteenReviewCommand(storeID int) CreateCanteenReviewCommand {
	return CreateCanteenReviewCommand{
		StoreID:         storeID,
		VisitedAt:       time.Now(),
		Taste:           3,
		Value:           4,
		Queue:           2,
		Hygiene:         5,
		Service:         1,
		Content:         "还不错",
		Tags:            []string{"上菜快"},
		ClientRequestID: "6f1a4c2e-8d3b-4f5a-9c7e-1b2d3e4f5a6b",
	}
}

func TestCanteenReviewCreate(t *testing.T) {
	service, _, _, _, _ := newCanteenReviewCommandFixture(t)
	user := &auth.User{ID: 2, Role: auth.RoleUser}
	review, created, err := service.Create(context.Background(), user, validCanteenReviewCommand(1))
	if err != nil {
		t.Fatalf("Create() = %v", err)
	}
	if !created || review.ID == 0 {
		t.Fatalf("created = %v, review = %+v", created, review)
	}
	if review.OverallScore != 3.25 {
		t.Fatalf("OverallScore = %v, want 3.25", review.OverallScore)
	}
	if review.Status != canteen.ReviewStatusActive || review.UserID != 2 {
		t.Fatalf("review = %+v", review)
	}
}

func TestCanteenReviewIdempotent(t *testing.T) {
	service, reviews, _, _, _ := newCanteenReviewCommandFixture(t)
	user := &auth.User{ID: 2, Role: auth.RoleUser}
	cmd := validCanteenReviewCommand(1)
	first, created, err := service.Create(context.Background(), user, cmd)
	if err != nil || !created {
		t.Fatalf("first Create() = %v, created=%v", err, created)
	}
	second, created, err := service.Create(context.Background(), user, cmd)
	if err != nil {
		t.Fatalf("second Create() = %v", err)
	}
	if created || second.ID != first.ID {
		t.Fatalf("second create = %+v created=%v", second, created)
	}
	if len(reviews.Reviews) != 1 {
		t.Fatalf("review count = %d, want 1", len(reviews.Reviews))
	}
}

func TestCanteenReviewRejectsInactiveStore(t *testing.T) {
	service, _, stores, _, _ := newCanteenReviewCommandFixture(t)
	stores.Stores[1].Status = canteen.StoreStatusOffline
	_, _, err := service.Create(context.Background(), &auth.User{ID: 2}, validCanteenReviewCommand(1))
	if !errors.Is(err, canteen.ErrStoreNotActive) {
		t.Fatalf("Create() err = %v, want ErrStoreNotActive", err)
	}
}

func TestCanteenReviewDailyLimit(t *testing.T) {
	service, _, _, _, query := newCanteenReviewCommandFixture(t)
	query.Count = 5
	_, _, err := service.Create(context.Background(), &auth.User{ID: 2}, validCanteenReviewCommand(1))
	if !errors.Is(err, canteen.ErrDailyReviewLimit) {
		t.Fatalf("Create() err = %v, want ErrDailyReviewLimit", err)
	}
}

func TestCanteenReviewFrequency(t *testing.T) {
	service, _, _, _, query := newCanteenReviewCommandFixture(t)
	service.maxDailyReviews = 100
	query.Count = 10
	_, _, err := service.Create(context.Background(), &auth.User{ID: 2}, validCanteenReviewCommand(1))
	if !errors.Is(err, canteen.ErrTooFrequent) {
		t.Fatalf("Create() err = %v, want ErrTooFrequent", err)
	}
}

func TestCanteenReviewDisabledToggle(t *testing.T) {
	stores := canteen.NewMockStoreRepository()
	reviews := canteen.NewMockReviewRepository()
	votes := canteen.NewMockVoteRepository()
	service := NewCanteenReviewCommandService(
		stores, reviews, votes, &canteen.MockReviewQuery{},
		StaticCanteenSettingsProvider{Runtime: CanteenRuntimeConfig{Enabled: true, ReviewsEnabled: false}},
		policy.Config{}, 5, nil,
	)
	_, _, err := service.Create(context.Background(), &auth.User{ID: 2}, validCanteenReviewCommand(1))
	if err == nil {
		t.Fatal("Create() should fail when reviews are disabled")
	}
}

func TestCanteenReviewUpdateAndConflict(t *testing.T) {
	service, reviews, _, _, _ := newCanteenReviewCommandFixture(t)
	user := &auth.User{ID: 2, Role: auth.RoleUser}
	created, _, err := service.Create(context.Background(), user, validCanteenReviewCommand(1))
	if err != nil {
		t.Fatalf("seed review: %v", err)
	}

	cmd := UpdateCanteenReviewCommand{
		ReviewID:  created.ID,
		VisitedAt: time.Now(),
		Taste:     5, Value: 5, Queue: 5, Hygiene: 5, Service: 5,
		Content:           "第二次来好多了",
		ExpectedUpdatedAt: created.UpdatedAt,
	}
	updated, err := service.Update(context.Background(), user, cmd)
	if err != nil {
		t.Fatalf("Update() = %v", err)
	}
	if updated.OverallScore != 5 {
		t.Fatalf("OverallScore = %v, want 5", updated.OverallScore)
	}

	second := cmd
	second.ExpectedUpdatedAt = updated.UpdatedAt
	second.Content = "第三次来"
	updatedAgain, err := service.Update(context.Background(), user, second)
	if err != nil {
		t.Fatalf("second Update() = %v", err)
	}
	if !updatedAgain.UpdatedAt.After(updated.UpdatedAt) {
		t.Fatalf("updated_at did not advance: %v -> %v", updated.UpdatedAt, updatedAgain.UpdatedAt)
	}

	missingVersion := cmd
	missingVersion.ExpectedUpdatedAt = time.Time{}
	if _, err := service.Update(context.Background(), user, missingVersion); err == nil {
		t.Fatal("Update() without version should be rejected")
	}

	cmd.ExpectedUpdatedAt = created.UpdatedAt
	if _, err := service.Update(context.Background(), user, cmd); !errors.Is(err, canteen.ErrReviewConflict) {
		t.Fatalf("Update(stale) err = %v, want ErrReviewConflict", err)
	}

	other := &auth.User{ID: 3, Role: auth.RoleUser}
	if _, err := service.Update(context.Background(), other, cmd); !errors.Is(err, canteen.ErrCannotEditReview) {
		t.Fatalf("Update(other) err = %v, want ErrCannotEditReview", err)
	}
	_ = reviews
}

func TestCanteenReviewDeleteAndVote(t *testing.T) {
	service, _, _, votes, _ := newCanteenReviewCommandFixture(t)
	author := &auth.User{ID: 2, Role: auth.RoleUser}
	review, _, err := service.Create(context.Background(), author, validCanteenReviewCommand(1))
	if err != nil {
		t.Fatalf("seed review: %v", err)
	}

	if err := service.Vote(context.Background(), author, review.ID, int(canteen.VoteLike)); !errors.Is(err, canteen.ErrVoteOwnReview) {
		t.Fatalf("Vote(self) err = %v, want ErrVoteOwnReview", err)
	}
	voter := &auth.User{ID: 3, Role: auth.RoleUser}
	if err := service.Vote(context.Background(), voter, review.ID, int(canteen.VoteLike)); err != nil {
		t.Fatalf("Vote() = %v", err)
	}
	if len(votes.Votes) != 1 {
		t.Fatalf("votes = %+v", votes.Votes)
	}
	if err := service.Vote(context.Background(), voter, review.ID, 0); err != nil {
		t.Fatalf("Vote(cancel) = %v", err)
	}
	if len(votes.Votes) != 0 {
		t.Fatalf("votes after cancel = %+v", votes.Votes)
	}
	if err := service.Delete(context.Background(), author, review.ID); err != nil {
		t.Fatalf("Delete() = %v", err)
	}
}

func TestCanteenReviewHideRequiresAdmin(t *testing.T) {
	service, _, _, _, _ := newCanteenReviewCommandFixture(t)
	author := &auth.User{ID: 2, Role: auth.RoleUser}
	review, _, err := service.Create(context.Background(), author, validCanteenReviewCommand(1))
	if err != nil {
		t.Fatalf("seed review: %v", err)
	}
	if err := service.Hide(context.Background(), author, review.ID, "spam"); !errors.Is(err, canteen.ErrCannotModerate) {
		t.Fatalf("Hide(author) err = %v, want ErrCannotModerate", err)
	}
	admin := &auth.User{ID: 1, Role: auth.RoleAdmin}
	if err := service.Hide(context.Background(), admin, review.ID, "spam"); err != nil {
		t.Fatalf("Hide(admin) = %v", err)
	}
}

func TestCanteenReviewVisitTimeBounds(t *testing.T) {
	service, _, _, _, _ := newCanteenReviewCommandFixture(t)
	user := &auth.User{ID: 2, Role: auth.RoleUser}

	future := validCanteenReviewCommand(1)
	future.VisitedAt = time.Now().Add(48 * time.Hour)
	if _, _, err := service.Create(context.Background(), user, future); err == nil {
		t.Fatal("Create() should reject a far-future visit time")
	}
	old := validCanteenReviewCommand(1)
	old.ClientRequestID = "6f1a4c2e-8d3b-4f5a-9c7e-1b2d3e4f5a6c"
	old.VisitedAt = time.Now().AddDate(-2, 0, 0)
	if _, _, err := service.Create(context.Background(), user, old); err == nil {
		t.Fatal("Create() should reject a visit older than one year")
	}
}
