package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/canteen"
	"jcourse/internal/infrastructure/repository"
)

func newCanteenDish(t *testing.T, dishes *repository.CanteenDishRepository, storeID int, name string, status canteen.DishStatus) *canteen.Dish {
	t.Helper()
	dish := &canteen.Dish{StoreID: storeID, Name: name, Status: status}
	dish.ApplyDefaults()
	if err := dishes.Create(context.Background(), dish); err != nil {
		t.Fatalf("create dish: %v", err)
	}
	return dish
}

func createCanteenDishReview(t *testing.T, reviews *repository.CanteenDishReviewRepository, dishID, userID int, score int) *canteen.DishReviewEvent {
	t.Helper()
	review := &canteen.DishReviewEvent{
		DishID:         dishID,
		UserID:         userID,
		DishDimensions: canteen.DishDimensions{Taste: score, Portion: score, Value: score},
		Content:        "菜品评价",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	review.ApplyCreate(canteen.DefaultDishWeights)
	if err := reviews.Create(context.Background(), review); err != nil {
		t.Fatalf("create dish review: %v", err)
	}
	return review
}

func TestCanteenDishAggregatesAndMerge(t *testing.T) {
	db := newTestDB(t)
	cleanTables(
		t, db,
		"canteen_review_dishes", "canteen_review_events", "canteen_review_revisions",
		"canteen_review_votes", "canteen_user_rating_summaries",
		"canteen_dish_review_events", "canteen_dish_rating_summaries", "canteen_dishes",
		"canteen_stores", "canteen_halls",
	)
	ctx := context.Background()

	halls := repository.NewCanteenHallRepository(db)
	stores := repository.NewCanteenStoreRepository(db)
	dishes := repository.NewCanteenDishRepository(db)
	dishReviews := repository.NewCanteenDishReviewRepository(db)
	dishQuery := repository.NewCanteenDishQueryRepository(db)
	reviews := repository.NewCanteenReviewRepository(db)

	hall := newCanteenHall(t, halls)
	store := newCanteenStore(t, stores, hall.ID, "老王面馆", canteen.StoreStatusActive)
	dish := newCanteenDish(t, dishes, store.ID, "牛肉面", canteen.DishStatusActive)

	createCanteenDishReview(t, dishReviews, dish.ID, 2, 5)
	createCanteenDishReview(t, dishReviews, dish.ID, 2, 1)

	reloaded, err := dishes.Get(ctx, dish.ID)
	if err != nil {
		t.Fatalf("get dish: %v", err)
	}
	if reloaded.Stats.RatingCount != 1 || reloaded.Stats.ReviewEventCount != 2 {
		t.Fatalf("dish stats = %+v", reloaded.Stats)
	}
	if reloaded.Stats.RatingAvg != 2.33 {
		t.Fatalf("dish rating avg = %v, want 2.33", reloaded.Stats.RatingAvg)
	}

	createCanteenDishReview(t, dishReviews, dish.ID, 3, 4)
	reloaded, err = dishes.Get(ctx, dish.ID)
	if err != nil {
		t.Fatalf("get dish: %v", err)
	}
	if reloaded.Stats.RatingCount != 2 || reloaded.Stats.ReviewEventCount != 3 {
		t.Fatalf("dish stats after second user = %+v", reloaded.Stats)
	}

	second := newCanteenDish(t, dishes, store.ID, "牛肉面（旧）", canteen.DishStatusActive)
	createCanteenDishReview(t, dishReviews, second.ID, 4, 3)
	visit := createCanteenReview(t, reviews, store.ID, 5, 5, time.Now())
	if err := db.Exec(
		"INSERT INTO canteen_review_dishes (review_event_id, dish_id) VALUES (?, ?)",
		visit.ID, second.ID,
	).Error; err != nil {
		t.Fatalf("link review dish: %v", err)
	}
	if err := dishes.Merge(ctx, second.ID, dish.ID, time.Now()); err != nil {
		t.Fatalf("merge dish: %v", err)
	}

	var linkedDishID int
	if err := db.Raw(
		"SELECT dish_id FROM canteen_review_dishes WHERE review_event_id = ?", visit.ID,
	).Scan(&linkedDishID).Error; err != nil {
		t.Fatalf("read review dish link: %v", err)
	}
	if linkedDishID != dish.ID {
		t.Fatalf("review dish link = %d, want %d", linkedDishID, dish.ID)
	}

	merged, err := dishes.Get(ctx, second.ID)
	if err != nil {
		t.Fatalf("get merged dish: %v", err)
	}
	if merged.Status != canteen.DishStatusMerged || merged.MergedIntoDishID == nil || *merged.MergedIntoDishID != dish.ID {
		t.Fatalf("merged dish = %+v", merged)
	}
	target, err := dishes.Get(ctx, dish.ID)
	if err != nil {
		t.Fatalf("get target dish: %v", err)
	}
	if target.Stats.RatingCount != 3 || target.Stats.ReviewEventCount != 4 {
		t.Fatalf("target stats after merge = %+v", target.Stats)
	}

	items, err := dishQuery.ListDishes(ctx, canteen.DishListFilter{StoreID: store.ID, Limit: 10})
	if err != nil {
		t.Fatalf("list dishes: %v", err)
	}
	if len(items) != 1 || items[0].ID != dish.ID {
		t.Fatalf("dishes = %+v", items)
	}
}
