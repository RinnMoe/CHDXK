package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/canteen"
	"jcourse/internal/infrastructure/repository"
)

func TestCanteenStoreQuery(t *testing.T) {
	db := newTestDB(t)
	cleanTables(
		t, db,
		"canteen_review_events", "canteen_review_revisions", "canteen_review_votes",
		"canteen_user_rating_summaries", "canteen_review_dishes", "canteen_dishes",
		"canteen_stores", "canteen_halls",
	)
	ctx := context.Background()
	now := time.Now()

	halls := repository.NewCanteenHallRepository(db)
	stores := repository.NewCanteenStoreRepository(db)
	reviews := repository.NewCanteenReviewRepository(db)
	query := repository.NewCanteenStoreQueryRepository(db)

	hall := newCanteenHall(t, halls)
	store1 := newCanteenStore(t, stores, hall.ID, "老王面馆", canteen.StoreStatusActive)
	store2 := &canteen.Store{
		HallID:     hall.ID,
		Name:       "一食堂快餐",
		Status:     canteen.StoreStatusActive,
		Categories: []string{"快餐"},
	}
	store2.ApplyDefaults()
	if err := stores.Create(ctx, store2); err != nil {
		t.Fatalf("create store2: %v", err)
	}
	newCanteenStore(t, stores, hall.ID, "待审核窗口", canteen.StoreStatusPending)
	newCanteenStore(t, stores, hall.ID, "停业窗口", canteen.StoreStatusOffline)

	if err := stores.UpdateStats(ctx, store1.ID, canteen.StoreStats{
		RatingCount: 2, RatingAvg: 4.5, RatingScore: 4.2,
		TasteAvg: 4.5, ValueAvg: 4, QueueAvg: 3.5, HygieneAvg: 4, ServiceAvg: 4.5,
		ReviewEventCount: 3, LastReviewAt: &now, RatingUpdatedAt: &now,
	}, now); err != nil {
		t.Fatalf("update store1 stats: %v", err)
	}
	if err := stores.UpdateStats(ctx, store2.ID, canteen.StoreStats{
		RatingCount: 1, RatingAvg: 3.5, RatingScore: 3.2,
		TasteAvg: 3.5, ValueAvg: 3.5, QueueAvg: 3.5, HygieneAvg: 3.5, ServiceAvg: 3.5,
		ReviewEventCount: 1, LastReviewAt: &now, RatingUpdatedAt: &now,
	}, now); err != nil {
		t.Fatalf("update store2 stats: %v", err)
	}

	if err := db.Create(&repository.CanteenUserRatingSummaryEntity{
		StoreID: store1.ID, UserID: 2, WeightedScore: 4.5, WindowSize: 1,
		LifetimeEventCount: 1, ComputedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed summary: %v", err)
	}

	review := &canteen.ReviewEvent{
		StoreID:    store1.ID,
		UserID:     2,
		VisitedAt:  now,
		Dimensions: canteen.Dimensions{Taste: 5, Value: 4, Queue: 3, Hygiene: 4, Service: 5},
		Content:    "还不错",
		Tags:       []string{"上菜快"},
	}
	review.ApplyCreate(canteen.DefaultStoreWeights)
	if err := reviews.Create(ctx, review); err != nil {
		t.Fatalf("create review: %v", err)
	}

	hallsView, err := query.ListHalls(ctx)
	if err != nil {
		t.Fatalf("ListHalls() = %v", err)
	}
	if len(hallsView) != 1 || hallsView[0].StoreCount != 2 {
		t.Fatalf("halls = %+v", hallsView)
	}

	items, total, err := query.ListStores(ctx, canteen.StoreListFilter{})
	if err != nil {
		t.Fatalf("ListStores() = %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("active stores = %d/%d, want 2", total, len(items))
	}
	if items[0].RatingScore == 0 || items[0].HallName != "一食堂" {
		t.Fatalf("store item = %+v", items[0])
	}

	filtered, filteredTotal, err := query.ListStores(ctx, canteen.StoreListFilter{Q: "老王"})
	if err != nil {
		t.Fatalf("ListStores(q) = %v", err)
	}
	if filteredTotal != 1 || len(filtered) != 1 || filtered[0].ID != store1.ID {
		t.Fatalf("filtered = %+v total=%d", filtered, filteredTotal)
	}

	byCategory, categoryTotal, err := query.ListStores(ctx, canteen.StoreListFilter{Category: "面食"})
	if err != nil {
		t.Fatalf("ListStores(category) = %v", err)
	}
	if categoryTotal != 1 || len(byCategory) != 1 {
		t.Fatalf("category stores = %+v total=%d", byCategory, categoryTotal)
	}

	ranked, rankedTotal, err := query.ListRankings(ctx, canteen.RankingFilter{Sort: "rating_score", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListRankings() = %v", err)
	}
	if rankedTotal != 2 || len(ranked) != 2 || ranked[0].ID != store1.ID {
		t.Fatalf("rankings = %+v total=%d", ranked, rankedTotal)
	}

	detail, err := query.GetStoreDetail(ctx, store1.ID, 2)
	if err != nil {
		t.Fatalf("GetStoreDetail() = %v", err)
	}
	if detail == nil || detail.Store == nil {
		t.Fatalf("detail = %+v", detail)
	}
	if detail.Dimensions.Taste != 5 || detail.MySummary == nil || detail.MyLatestReview == nil {
		t.Fatalf("detail = %+v", detail)
	}
	if detail.MySummary.WeightedScore != 4.3 || detail.MyLatestReview.Content != "还不错" {
		t.Fatalf("my summary = %+v, latest = %+v", detail.MySummary, detail.MyLatestReview)
	}

	home, err := query.GetHome(ctx)
	if err != nil {
		t.Fatalf("GetHome() = %v", err)
	}
	if home.Stats.StoreCount != 2 || home.Stats.ReviewCount != 1 {
		t.Fatalf("home stats = %+v", home.Stats)
	}
	if len(home.TopStores) != 2 || len(home.NewStores) != 2 {
		t.Fatalf("home stores = %d/%d", len(home.TopStores), len(home.NewStores))
	}
	if len(home.LatestReviews) != 1 || home.LatestReviews[0].StoreName != "老王面馆" {
		t.Fatalf("home reviews = %+v", home.LatestReviews)
	}
}
