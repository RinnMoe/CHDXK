package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/canteen"
	"jcourse/internal/infrastructure/repository"
)

func TestCanteenStoreMergeMovesReviewsAndDishes(t *testing.T) {
	db := newTestDB(t)
	cleanTables(
		t, db,
		"canteen_review_dishes", "canteen_dish_media", "canteen_dish_review_events",
		"canteen_dish_rating_summaries", "canteen_dishes", "canteen_review_media",
		"canteen_media", "canteen_review_events", "canteen_user_rating_summaries",
		"canteen_stores", "canteen_halls",
	)
	ctx := context.Background()
	halls := repository.NewCanteenHallRepository(db)
	stores := repository.NewCanteenStoreRepository(db)
	dishes := repository.NewCanteenDishRepository(db)
	reviews := repository.NewCanteenReviewRepository(db)
	dishReviews := repository.NewCanteenDishReviewRepository(db)
	query := repository.NewCanteenStoreQueryRepository(db)

	hall := newCanteenHall(t, halls)
	source := newCanteenStore(t, stores, hall.ID, "老王面馆", canteen.StoreStatusActive)
	target := newCanteenStore(t, stores, hall.ID, "老王面馆(新)", canteen.StoreStatusActive)
	sourceDish := newCanteenDish(t, dishes, source.ID, "牛肉面", canteen.DishStatusActive)
	targetDish := newCanteenDish(t, dishes, target.ID, "牛肉面", canteen.DishStatusActive)

	base := time.Now().Add(-time.Hour)
	createTaggedReview(t, reviews, source.ID, 2, 5, []string{"上菜快"}, base)
	createTaggedReview(t, reviews, target.ID, 3, 4, []string{"环境干净"}, base.Add(time.Minute))
	createCanteenDishReview(t, dishReviews, sourceDish.ID, 4, 5)

	if err := stores.Merge(ctx, source.ID, target.ID, time.Now()); err != nil {
		t.Fatalf("merge stores: %v", err)
	}

	mergedSource, err := stores.Get(ctx, source.ID)
	if err != nil {
		t.Fatalf("get source: %v", err)
	}
	if mergedSource.Status != canteen.StoreStatusMerged || mergedSource.MergedIntoStoreID == nil || *mergedSource.MergedIntoStoreID != target.ID {
		t.Fatalf("source = %+v", mergedSource)
	}
	if mergedSource.Stats.ReviewEventCount != 0 || mergedSource.Stats.RatingCount != 0 {
		t.Fatalf("source stats = %+v", mergedSource.Stats)
	}

	mergedTarget, err := stores.Get(ctx, target.ID)
	if err != nil {
		t.Fatalf("get target: %v", err)
	}
	if mergedTarget.Stats.ReviewEventCount != 2 || mergedTarget.Stats.RatingCount != 2 {
		t.Fatalf("target stats = %+v", mergedTarget.Stats)
	}

	targetReviews, total, err := query.ListReviews(ctx, canteen.ReviewListFilter{
		StoreID: target.ID, Order: "latest", History: true, PageSize: 10,
	})
	if err != nil || total != 2 || len(targetReviews) != 2 {
		t.Fatalf("target reviews = %+v total=%d err=%v", targetReviews, total, err)
	}
	sourceReviews, sourceTotal, err := query.ListReviews(ctx, canteen.ReviewListFilter{
		StoreID: source.ID, Order: "latest", History: true, PageSize: 10,
	})
	if err != nil || sourceTotal != 0 || len(sourceReviews) != 0 {
		t.Fatalf("source reviews = %+v total=%d err=%v", sourceReviews, sourceTotal, err)
	}

	mergedDish, err := dishes.Get(ctx, sourceDish.ID)
	if err != nil {
		t.Fatalf("get source dish: %v", err)
	}
	if mergedDish.Status != canteen.DishStatusMerged || mergedDish.MergedIntoDishID == nil || *mergedDish.MergedIntoDishID != targetDish.ID {
		t.Fatalf("source dish = %+v", mergedDish)
	}
	targetDishReloaded, err := dishes.Get(ctx, targetDish.ID)
	if err != nil {
		t.Fatalf("get target dish: %v", err)
	}
	if targetDishReloaded.Stats.RatingCount != 1 || targetDishReloaded.Stats.ReviewEventCount != 1 {
		t.Fatalf("target dish stats = %+v", targetDishReloaded.Stats)
	}
}

func TestCanteenStoreMergeSameUserCountsOnce(t *testing.T) {
	db := newTestDB(t)
	cleanTables(
		t, db,
		"canteen_review_events", "canteen_review_revisions", "canteen_review_votes",
		"canteen_user_rating_summaries", "canteen_stores", "canteen_halls",
	)
	ctx := context.Background()
	halls := repository.NewCanteenHallRepository(db)
	stores := repository.NewCanteenStoreRepository(db)
	reviews := repository.NewCanteenReviewRepository(db)
	query := repository.NewCanteenStoreQueryRepository(db)

	hall := newCanteenHall(t, halls)
	source := newCanteenStore(t, stores, hall.ID, "同用户源店", canteen.StoreStatusActive)
	target := newCanteenStore(t, stores, hall.ID, "同用户目标店", canteen.StoreStatusActive)
	base := time.Now().Add(-2 * time.Hour)
	createTaggedReview(t, reviews, source.ID, 2, 5, []string{"上菜快"}, base)
	createTaggedReview(t, reviews, target.ID, 2, 3, []string{"环境干净"}, base.Add(time.Hour))

	if err := stores.Merge(ctx, source.ID, target.ID, time.Now()); err != nil {
		t.Fatalf("merge stores: %v", err)
	}
	merged, err := stores.Get(ctx, target.ID)
	if err != nil {
		t.Fatalf("get target: %v", err)
	}
	if merged.Stats.RatingCount != 1 || merged.Stats.ReviewEventCount != 2 {
		t.Fatalf("target stats = %+v, want one user/two events", merged.Stats)
	}
	latest, latestTotal, err := query.ListReviews(ctx, canteen.ReviewListFilter{
		StoreID: target.ID, Order: "latest", PageSize: 10,
	})
	if err != nil || latestTotal != 1 || len(latest) != 1 {
		t.Fatalf("latest total=%d items=%d err=%v", latestTotal, len(latest), err)
	}
	history, historyTotal, err := query.ListReviews(ctx, canteen.ReviewListFilter{
		StoreID: target.ID, Order: "latest", History: true, PageSize: 10,
	})
	if err != nil || historyTotal != 2 || len(history) != 2 {
		t.Fatalf("history total=%d items=%d err=%v", historyTotal, len(history), err)
	}
}
