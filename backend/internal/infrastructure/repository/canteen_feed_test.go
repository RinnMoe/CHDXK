package repository_test

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/canteen"
	"jcourse/internal/infrastructure/repository"
)

func createTaggedReview(t *testing.T, reviews *repository.CanteenReviewRepository, storeID, userID, score int, tags []string, visitedAt time.Time) *canteen.ReviewEvent {
	t.Helper()
	review := &canteen.ReviewEvent{
		StoreID:    storeID,
		UserID:     userID,
		VisitedAt:  visitedAt,
		Dimensions: canteen.Dimensions{Taste: score, Value: score, Queue: score, Hygiene: score, Service: score},
		Content:    "测试评价",
		Tags:       tags,
	}
	review.ApplyCreate(canteen.DefaultStoreWeights)
	if err := reviews.Create(context.Background(), review); err != nil {
		t.Fatalf("create review: %v", err)
	}
	return review
}

func TestCanteenReviewLatestHistoryFiltersAndCursor(t *testing.T) {
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
	query := repository.NewCanteenStoreQueryRepository(db)

	hall := newCanteenHall(t, halls)
	store := newCanteenStore(t, stores, hall.ID, "老王面馆", canteen.StoreStatusActive)
	base := time.Now().Add(-time.Hour)
	first := createTaggedReview(t, reviews, store.ID, 2, 5, []string{"上菜快"}, base)
	createTaggedReview(t, reviews, store.ID, 2, 2, []string{"等位久"}, base.Add(10*time.Minute))
	createTaggedReview(t, reviews, store.ID, 3, 4, []string{"环境干净"}, base.Add(20*time.Minute))

	latest, total, err := query.ListReviews(ctx, canteen.ReviewListFilter{StoreID: store.ID, Order: "latest", PageSize: 10})
	if err != nil {
		t.Fatalf("list latest: %v", err)
	}
	if total != 2 || len(latest) != 2 {
		t.Fatalf("latest total=%d items=%d, want 2", total, len(latest))
	}
	history, historyTotal, err := query.ListReviews(ctx, canteen.ReviewListFilter{StoreID: store.ID, Order: "latest", History: true, PageSize: 10})
	if err != nil {
		t.Fatalf("list history: %v", err)
	}
	if historyTotal != 3 || len(history) != 3 {
		t.Fatalf("history total=%d items=%d, want 3", historyTotal, len(history))
	}

	high := 4.0
	highItems, _, err := query.ListReviews(ctx, canteen.ReviewListFilter{StoreID: store.ID, Order: "latest", MinScore: &high, PageSize: 10})
	if err != nil {
		t.Fatalf("list high: %v", err)
	}
	if len(highItems) != 1 || highItems[0].UserID != 3 {
		t.Fatalf("high items = %+v", highItems)
	}

	pageOne, _, err := query.ListReviews(ctx, canteen.ReviewListFilter{StoreID: store.ID, Order: "latest", History: true, PageSize: 1})
	if err != nil || len(pageOne) != 1 {
		t.Fatalf("page one = %+v err=%v", pageOne, err)
	}
	firstItem := pageOne[0]
	cursor := &canteen.ReviewCursor{
		Order: "latest", OverallScore: firstItem.OverallScore,
		LikeCount: firstItem.LikeCount, CreatedAt: firstItem.CreatedAt, ID: firstItem.ID,
	}
	pageTwo, _, err := query.ListReviews(ctx, canteen.ReviewListFilter{
		StoreID: store.ID, Order: "latest", History: true, PageSize: 1, Cursor: cursor,
	})
	if err != nil || len(pageTwo) != 1 {
		t.Fatalf("page two = %+v err=%v", pageTwo, err)
	}
	if pageTwo[0].ID == firstItem.ID {
		t.Fatalf("cursor returned duplicate id %d", firstItem.ID)
	}
	if firstItem.ID == first.ID {
		t.Fatalf("latest page should not return the oldest event")
	}
}

func TestCanteenRankingExplanationAndDishGallery(t *testing.T) {
	db := newTestDB(t)
	cleanTables(
		t, db,
		"canteen_dish_media", "canteen_dish_review_events", "canteen_dish_rating_summaries",
		"canteen_dishes", "canteen_review_events", "canteen_review_media", "canteen_media",
		"canteen_user_rating_summaries", "canteen_stores", "canteen_halls",
	)
	ctx := context.Background()
	halls := repository.NewCanteenHallRepository(db)
	stores := repository.NewCanteenStoreRepository(db)
	reviews := repository.NewCanteenReviewRepository(db)
	dishes := repository.NewCanteenDishRepository(db)
	dishReviews := repository.NewCanteenDishReviewRepository(db)
	mediaRepo := repository.NewCanteenMediaRepository(db)
	query := repository.NewCanteenStoreQueryRepository(db)
	dishQuery := repository.NewCanteenDishQueryRepository(db)

	hall := newCanteenHall(t, halls)
	store := newCanteenStore(t, stores, hall.ID, "老王面馆", canteen.StoreStatusActive)
	base := time.Now().Add(-time.Hour)
	createTaggedReview(t, reviews, store.ID, 2, 5, []string{"上菜快", "环境干净"}, base)
	createTaggedReview(t, reviews, store.ID, 3, 4, []string{"上菜快", "环境干净"}, base.Add(time.Minute))

	dish := newCanteenDish(t, dishes, store.ID, "牛肉面", canteen.DishStatusActive)
	createCanteenDishReview(t, dishReviews, dish.ID, 2, 5)
	now := time.Now()
	media := &canteen.Media{
		OwnerUserID: 2, Purpose: canteen.MediaPurposeDish, StorageDriver: "local",
		StorageKey: "dish/test.png", MIME: "image/png", ByteSize: 10, Width: 8, Height: 8,
		SHA256: "abc", Status: canteen.MediaStatusApproved, CreatedAt: now, UpdatedAt: now,
	}
	if err := mediaRepo.Create(ctx, media); err != nil {
		t.Fatalf("create media: %v", err)
	}
	if err := mediaRepo.AttachToDish(ctx, dish.ID, media.ID, 2, 0, now); err != nil {
		t.Fatalf("attach dish media: %v", err)
	}

	rankings, _, err := query.ListRankings(ctx, canteen.RankingFilter{Sort: "rating_score", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("rankings: %v", err)
	}
	if len(rankings) != 1 {
		t.Fatalf("rankings = %+v", rankings)
	}
	item := rankings[0]
	if item.Confidence != "样本较少" {
		t.Fatalf("confidence = %q, want 样本较少", item.Confidence)
	}
	if item.DishCount != 1 || item.DishPhotoCount != 1 {
		t.Fatalf("dish counts = %d/%d", item.DishCount, item.DishPhotoCount)
	}
	if len(item.SummaryTags) == 0 || item.SummaryTags[0] != "上菜快" {
		t.Fatalf("summary tags = %+v", item.SummaryTags)
	}

	dishesList, err := dishQuery.ListDishes(ctx, canteen.DishListFilter{StoreID: store.ID, Limit: 10})
	if err != nil || len(dishesList) != 1 {
		t.Fatalf("dishes = %+v err=%v", dishesList, err)
	}
	if dishesList[0].PhotoCount != 1 || dishesList[0].CoverURL == "" || dishesList[0].StoreName != store.Name {
		t.Fatalf("dish view = %+v", dishesList[0])
	}
	photos, err := dishQuery.ListDishPhotos(ctx, dish.ID)
	if err != nil || len(photos) != 1 || photos[0].URL == "" {
		t.Fatalf("photos = %+v err=%v", photos, err)
	}
	dishReviewItems, dishTotal, err := dishQuery.ListDishReviews(ctx, dish.ID, 1, 10)
	if err != nil || dishTotal != 1 || len(dishReviewItems) != 1 {
		t.Fatalf("dish reviews = %+v total=%d err=%v", dishReviewItems, dishTotal, err)
	}

	home, err := query.GetHome(ctx)
	if err != nil {
		t.Fatalf("home: %v", err)
	}
	if home.Hero == nil || len(home.HotDishes) == 0 || len(home.Feeds) == 0 {
		t.Fatalf("home hero=%+v hot=%d feeds=%d", home.Hero, len(home.HotDishes), len(home.Feeds))
	}
}

func TestCanteenMediaVisibilityFollowsParentStatus(t *testing.T) {
	db := newTestDB(t)
	cleanTables(
		t, db,
		"canteen_dish_media", "canteen_dishes", "canteen_review_media", "canteen_media",
		"canteen_review_events", "canteen_user_rating_summaries", "canteen_stores", "canteen_halls",
	)
	ctx := context.Background()
	halls := repository.NewCanteenHallRepository(db)
	stores := repository.NewCanteenStoreRepository(db)
	reviews := repository.NewCanteenReviewRepository(db)
	dishes := repository.NewCanteenDishRepository(db)
	mediaRepo := repository.NewCanteenMediaRepository(db)

	hall := newCanteenHall(t, halls)
	store := newCanteenStore(t, stores, hall.ID, "媒体可见性店", canteen.StoreStatusActive)
	review := createTaggedReview(t, reviews, store.ID, 2, 5, nil, time.Now().Add(-time.Hour))
	media := &canteen.Media{
		OwnerUserID: 2, Purpose: canteen.MediaPurposeReview, StorageDriver: "local",
		StorageKey: "review/visibility.png", MIME: "image/png", ByteSize: 10, Width: 8, Height: 8,
		SHA256: "visibility", Status: canteen.MediaStatusApproved, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := mediaRepo.Create(ctx, media); err != nil {
		t.Fatalf("create media: %v", err)
	}
	if err := mediaRepo.AttachToReview(ctx, review.ID, media.ID, 0); err != nil {
		t.Fatalf("attach review media: %v", err)
	}
	visible, err := mediaRepo.IsMediaPublic(ctx, media.ID)
	if err != nil || !visible {
		t.Fatalf("review media visible=%v err=%v", visible, err)
	}
	if err := reviews.Hide(ctx, review.ID, 1, "test", time.Now()); err != nil {
		t.Fatalf("hide review: %v", err)
	}
	visible, err = mediaRepo.IsMediaPublic(ctx, media.ID)
	if err != nil || visible {
		t.Fatalf("hidden review media visible=%v err=%v", visible, err)
	}

	dish := newCanteenDish(t, dishes, store.ID, "可见性菜品", canteen.DishStatusActive)
	dishMedia := &canteen.Media{
		OwnerUserID: 2, Purpose: canteen.MediaPurposeDish, StorageDriver: "local",
		StorageKey: "dish/visibility.png", MIME: "image/png", ByteSize: 10, Width: 8, Height: 8,
		SHA256: "visibility-dish", Status: canteen.MediaStatusApproved, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := mediaRepo.Create(ctx, dishMedia); err != nil {
		t.Fatalf("create dish media: %v", err)
	}
	if err := mediaRepo.AttachToDish(ctx, dish.ID, dishMedia.ID, 2, 0, time.Now()); err != nil {
		t.Fatalf("attach dish media: %v", err)
	}
	visible, err = mediaRepo.IsMediaPublic(ctx, dishMedia.ID)
	if err != nil || !visible {
		t.Fatalf("dish media visible=%v err=%v", visible, err)
	}
	if err := db.Exec("UPDATE canteen_stores SET status = ? WHERE id = ?", string(canteen.StoreStatusOffline), store.ID).Error; err != nil {
		t.Fatalf("offline store: %v", err)
	}
	visible, err = mediaRepo.IsMediaPublic(ctx, dishMedia.ID)
	if err != nil || !visible {
		t.Fatalf("offline store dish media visible=%v err=%v", visible, err)
	}
	if err := db.Exec("UPDATE canteen_stores SET status = ? WHERE id = ?", string(canteen.StoreStatusActive), store.ID).Error; err != nil {
		t.Fatalf("reactivate store: %v", err)
	}
	if err := dishes.UpdateStatus(ctx, dish.ID, canteen.DishStatusHidden, time.Now()); err != nil {
		t.Fatalf("hide dish: %v", err)
	}
	visible, err = mediaRepo.IsMediaPublic(ctx, dishMedia.ID)
	if err != nil || visible {
		t.Fatalf("hidden dish media visible=%v err=%v", visible, err)
	}

	cover := &canteen.Media{
		OwnerUserID: 2, Purpose: canteen.MediaPurposeStoreCover, StorageDriver: "local",
		StorageKey: "store/cover.png", MIME: "image/png", ByteSize: 10, Width: 8, Height: 8,
		SHA256: "visibility-cover", Status: canteen.MediaStatusApproved, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := mediaRepo.Create(ctx, cover); err != nil {
		t.Fatalf("create cover: %v", err)
	}
	if err := db.Exec("UPDATE canteen_stores SET cover_media_id = ? WHERE id = ?", cover.ID, store.ID).Error; err != nil {
		t.Fatalf("attach cover: %v", err)
	}
	visible, err = mediaRepo.IsMediaPublic(ctx, cover.ID)
	if err != nil || !visible {
		t.Fatalf("store cover visible=%v err=%v", visible, err)
	}
	if err := db.Exec("UPDATE canteen_stores SET status = ? WHERE id = ?", string(canteen.StoreStatusOffline), store.ID).Error; err != nil {
		t.Fatalf("offline cover store: %v", err)
	}
	visible, err = mediaRepo.IsMediaPublic(ctx, cover.ID)
	if err != nil || !visible {
		t.Fatalf("offline store cover visible=%v err=%v", visible, err)
	}
	if err := db.Exec("UPDATE canteen_stores SET status = ? WHERE id = ?", string(canteen.StoreStatusRejected), store.ID).Error; err != nil {
		t.Fatalf("reject cover store: %v", err)
	}
	visible, err = mediaRepo.IsMediaPublic(ctx, cover.ID)
	if err != nil || visible {
		t.Fatalf("rejected store cover visible=%v err=%v", visible, err)
	}
}
