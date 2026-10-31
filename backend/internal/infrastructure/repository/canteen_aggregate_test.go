package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/canteen"
	"jcourse/internal/infrastructure/repository"
)

func createCanteenReview(t *testing.T, reviews *repository.CanteenReviewRepository, storeID, userID int, score int, visitedAt time.Time) *canteen.ReviewEvent {
	t.Helper()
	review := &canteen.ReviewEvent{
		StoreID:    storeID,
		UserID:     userID,
		VisitedAt:  visitedAt,
		Dimensions: canteen.Dimensions{Taste: score, Value: score, Queue: score, Hygiene: score, Service: score},
		Content:    "测试评价",
	}
	review.ApplyCreate(canteen.DefaultStoreWeights)
	if err := reviews.Create(context.Background(), review); err != nil {
		t.Fatalf("create review: %v", err)
	}
	return review
}

func TestCanteenReviewAggregates(t *testing.T) {
	db := newTestDB(t)
	cleanTables(
		t, db,
		"canteen_review_events", "canteen_review_revisions", "canteen_review_votes",
		"canteen_user_rating_summaries", "canteen_dishes", "canteen_stores", "canteen_halls",
	)
	ctx := context.Background()

	halls := repository.NewCanteenHallRepository(db)
	stores := repository.NewCanteenStoreRepository(db)
	reviews := repository.NewCanteenReviewRepository(db)
	votes := repository.NewCanteenVoteRepository(db)
	query := repository.NewCanteenStoreQueryRepository(db)

	hall := newCanteenHall(t, halls)
	store := newCanteenStore(t, stores, hall.ID, "老王面馆", canteen.StoreStatusActive)

	base := time.Now().Add(-time.Hour)
	five := createCanteenReview(t, reviews, store.ID, 2, 5, base)
	one := createCanteenReview(t, reviews, store.ID, 2, 1, base.Add(time.Minute))

	summary, err := query.GetUserSummary(ctx, store.ID, 2)
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if summary == nil {
		t.Fatal("summary missing")
	}
	// newest first: 1 then 5 -> (1*0.6 + 5*0.3) / 0.9 = 2.33
	if summary.WeightedScore != 2.33 || summary.WindowSize != 2 {
		t.Fatalf("summary = %+v", summary)
	}

	reloaded, err := stores.Get(ctx, store.ID)
	if err != nil {
		t.Fatalf("get store: %v", err)
	}
	if reloaded.Stats.RatingCount != 1 {
		t.Fatalf("rating count = %d, want 1", reloaded.Stats.RatingCount)
	}
	if reloaded.Stats.RatingAvg != 2.33 {
		t.Fatalf("rating avg = %v, want 2.33", reloaded.Stats.RatingAvg)
	}
	if reloaded.Stats.ReviewEventCount != 2 {
		t.Fatalf("review count = %d, want 2", reloaded.Stats.ReviewEventCount)
	}

	// A third and fourth visit keep only the latest three events.
	createCanteenReview(t, reviews, store.ID, 2, 4, base.Add(2*time.Minute))
	createCanteenReview(t, reviews, store.ID, 2, 3, base.Add(3*time.Minute))
	summary, err = query.GetUserSummary(ctx, store.ID, 2)
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if summary.WindowSize != 3 || summary.LifetimeEventCount != 4 {
		t.Fatalf("summary = %+v", summary)
	}
	// latest three: 3,4,1 -> newest first weights 0.6/0.3/0.1 => (3*.6 + 4*.3 + 1*.1)/1 = 3.1
	if summary.WeightedScore != 3.1 {
		t.Fatalf("weighted score = %v, want 3.1", summary.WeightedScore)
	}

	// A second user becomes a second sample.
	createCanteenReview(t, reviews, store.ID, 3, 5, base.Add(4*time.Minute))
	reloaded, err = stores.Get(ctx, store.ID)
	if err != nil {
		t.Fatalf("get store: %v", err)
	}
	if reloaded.Stats.RatingCount != 2 {
		t.Fatalf("rating count = %d, want 2", reloaded.Stats.RatingCount)
	}

	// Voting updates the denormalized counters.
	if err := votes.Save(ctx, &canteen.Vote{
		ReviewEventID: five.ID, UserID: 3, VoteType: canteen.VoteLike,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("save vote: %v", err)
	}
	voted, err := reviews.Get(ctx, five.ID)
	if err != nil {
		t.Fatalf("get voted review: %v", err)
	}
	if voted.LikeCount != 1 {
		t.Fatalf("like count = %d, want 1", voted.LikeCount)
	}
	if err := votes.Delete(ctx, five.ID, 3); err != nil {
		t.Fatalf("delete vote: %v", err)
	}
	voted, err = reviews.Get(ctx, five.ID)
	if err != nil {
		t.Fatalf("get review: %v", err)
	}
	if voted.LikeCount != 0 {
		t.Fatalf("like count after delete = %d, want 0", voted.LikeCount)
	}

	// Deleting the newest event recomputes the summary and store stats.
	if err := reviews.SoftDelete(ctx, one.ID, time.Now()); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	summary, err = query.GetUserSummary(ctx, store.ID, 2)
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if summary.WindowSize != 3 || summary.LifetimeEventCount != 3 {
		t.Fatalf("summary after delete = %+v", summary)
	}
	reloaded, err = stores.Get(ctx, store.ID)
	if err != nil {
		t.Fatalf("get store: %v", err)
	}
	if reloaded.Stats.ReviewEventCount != 4 {
		t.Fatalf("store review count = %d, want 4", reloaded.Stats.ReviewEventCount)
	}
	if reloaded.Stats.RatingCount != 2 {
		t.Fatalf("store rating count = %d, want 2", reloaded.Stats.RatingCount)
	}
}
