package repository_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"jcourse/internal/domain/canteen"
	"jcourse/internal/infrastructure/repository"
	"jcourse/pkg/apperr"
)

func newCanteenHall(t *testing.T, halls *repository.CanteenHallRepository) *canteen.Hall {
	t.Helper()
	hall := &canteen.Hall{
		CampusKey: "benbu",
		Name:      "一食堂",
		Status:    canteen.HallStatusActive,
	}
	if err := halls.Create(context.Background(), hall); err != nil {
		t.Fatalf("create hall: %v", err)
	}
	return hall
}

func newCanteenStore(t *testing.T, stores *repository.CanteenStoreRepository, hallID int, name string, status canteen.StoreStatus) *canteen.Store {
	t.Helper()
	store := &canteen.Store{HallID: hallID, Name: name, Status: status, Categories: []string{"面食"}}
	store.ApplyDefaults()
	if err := stores.Create(context.Background(), store); err != nil {
		t.Fatalf("create store: %v", err)
	}
	return store
}

func TestCanteenStoreUniqueNamePerHall(t *testing.T) {
	db := newTestDB(t)
	cleanTables(t, db, "canteen_stores", "canteen_halls")
	ctx := context.Background()

	halls := repository.NewCanteenHallRepository(db)
	stores := repository.NewCanteenStoreRepository(db)
	hall := newCanteenHall(t, halls)
	store := newCanteenStore(t, stores, hall.ID, "老王面馆", canteen.StoreStatusActive)
	if store.ID == 0 || store.NormalizedName != "老王面馆" {
		t.Fatalf("store = %+v", store)
	}

	duplicate := &canteen.Store{HallID: hall.ID, Name: "老王面馆 ", Status: canteen.StoreStatusActive}
	duplicate.ApplyDefaults()
	err := stores.Create(ctx, duplicate)
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) || appErr.StatusCode != 409 {
		t.Fatalf("duplicate store err = %v, want 409", err)
	}

	found, err := stores.GetByHallAndNormalizedName(ctx, hall.ID, "老王面馆")
	if err != nil {
		t.Fatalf("find store: %v", err)
	}
	if found == nil || found.ID != store.ID {
		t.Fatalf("found = %+v", found)
	}

	// A merged store releases the name for reuse.
	target := newCanteenStore(t, stores, hall.ID, "老王面馆二店", canteen.StoreStatusActive)
	if err := stores.Merge(ctx, store.ID, target.ID, time.Now()); err != nil {
		t.Fatalf("merge store: %v", err)
	}
	reused := &canteen.Store{HallID: hall.ID, Name: "老王面馆", Status: canteen.StoreStatusActive}
	reused.ApplyDefaults()
	if err := stores.Create(ctx, reused); err != nil {
		t.Fatalf("reuse merged store name: %v", err)
	}
}

func TestCanteenReviewClientRequestIDIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	cleanTables(t, db, "canteen_review_events", "canteen_review_revisions", "canteen_stores", "canteen_halls")
	ctx := context.Background()

	halls := repository.NewCanteenHallRepository(db)
	stores := repository.NewCanteenStoreRepository(db)
	reviews := repository.NewCanteenReviewRepository(db)
	hall := newCanteenHall(t, halls)
	store := newCanteenStore(t, stores, hall.ID, "老王面馆", canteen.StoreStatusActive)

	review := &canteen.ReviewEvent{
		StoreID:         store.ID,
		UserID:          42,
		VisitedAt:       time.Now(),
		Dimensions:      canteen.Dimensions{Taste: 5, Value: 4, Queue: 3, Hygiene: 4, Service: 5},
		Content:         "还不错",
		ClientRequestID: "6f1a4c2e-8d3b-4f5a-9c7e-1b2d3e4f5a6b",
	}
	review.ApplyCreate(canteen.DefaultStoreWeights)
	if err := reviews.Create(ctx, review); err != nil {
		t.Fatalf("create review: %v", err)
	}

	same := *review
	same.ID = 0
	if err := reviews.Create(ctx, &same); !errors.Is(err, canteen.ErrReviewConflict) {
		t.Fatalf("duplicate request id err = %v, want ErrReviewConflict", err)
	}

	found, err := reviews.FindByClientRequestID(ctx, 42, review.ClientRequestID)
	if err != nil {
		t.Fatalf("find by request id: %v", err)
	}
	if found == nil || found.ID != review.ID {
		t.Fatalf("found = %+v", found)
	}

	// Another user may reuse the same client request id.
	otherUser := *review
	otherUser.ID = 0
	otherUser.UserID = 43
	if err := reviews.Create(ctx, &otherUser); err != nil {
		t.Fatalf("create review for other user: %v", err)
	}
}

func TestCanteenReviewUpdateWritesRevision(t *testing.T) {
	db := newTestDB(t)
	cleanTables(t, db, "canteen_review_events", "canteen_review_revisions", "canteen_stores", "canteen_halls")
	ctx := context.Background()

	halls := repository.NewCanteenHallRepository(db)
	stores := repository.NewCanteenStoreRepository(db)
	reviews := repository.NewCanteenReviewRepository(db)
	hall := newCanteenHall(t, halls)
	store := newCanteenStore(t, stores, hall.ID, "老王面馆", canteen.StoreStatusActive)

	review := &canteen.ReviewEvent{
		StoreID:    store.ID,
		UserID:     42,
		VisitedAt:  time.Now().Add(-time.Hour),
		Dimensions: canteen.Dimensions{Taste: 5, Value: 5, Queue: 5, Hygiene: 5, Service: 5},
	}
	review.ApplyCreate(canteen.DefaultStoreWeights)
	if err := reviews.Create(ctx, review); err != nil {
		t.Fatalf("create review: %v", err)
	}

	expectedUpdatedAt := review.UpdatedAt
	revision := review.MakeRevision(42, time.Now())
	review.ApplyUpdate(canteen.UpdateReview{
		Dimensions: canteen.Dimensions{Taste: 2, Value: 2, Queue: 2, Hygiene: 2, Service: 2},
		Content:    "第二次去变差了",
		Now:        time.Now(),
	}, canteen.DefaultStoreWeights)
	if err := reviews.Update(ctx, review, revision, expectedUpdatedAt); err != nil {
		t.Fatalf("update review: %v", err)
	}

	reloaded, err := reviews.Get(ctx, review.ID)
	if err != nil {
		t.Fatalf("get review: %v", err)
	}
	if reloaded.OverallScore != 2 || reloaded.Content != "第二次去变差了" {
		t.Fatalf("reloaded = %+v", reloaded)
	}

	var revisionCount int64
	if err := db.Model(&repository.CanteenReviewRevisionEntity{}).
		Where("review_event_id = ?", review.ID).Count(&revisionCount).Error; err != nil {
		t.Fatalf("count revisions: %v", err)
	}
	if revisionCount != 1 {
		t.Fatalf("revision count = %d, want 1", revisionCount)
	}

	if err := reviews.Hide(ctx, review.ID, 1, "内容不当", time.Now()); err != nil {
		t.Fatalf("hide review: %v", err)
	}
	hidden, err := reviews.Get(ctx, review.ID)
	if err != nil {
		t.Fatalf("get hidden review: %v", err)
	}
	if hidden.Status != canteen.ReviewStatusHidden || hidden.HiddenReason != "内容不当" {
		t.Fatalf("hidden = %+v", hidden)
	}

	if err := reviews.SoftDelete(ctx, review.ID, time.Now()); err != nil {
		t.Fatalf("delete review: %v", err)
	}
	deleted, err := reviews.Get(ctx, review.ID)
	if err != nil {
		t.Fatalf("get deleted review: %v", err)
	}
	if deleted.Status != canteen.ReviewStatusDeleted {
		t.Fatalf("deleted status = %q", deleted.Status)
	}
}

func TestCanteenPartialUniqueIndexes(t *testing.T) {
	db := newTestDB(t)
	var count int64
	if err := db.Raw(`
		SELECT count(*) FROM pg_indexes
		WHERE indexname IN (
			'uniq_canteen_stores_hall_name',
			'uniq_canteen_dishes_store_name',
			'uniq_canteen_review_events_user_request',
			'uniq_canteen_reports_open_subject'
		)`).Scan(&count).Error; err != nil {
		t.Fatalf("query indexes: %v", err)
	}
	if count != 4 {
		t.Fatalf("canteen partial index count = %d, want 4", count)
	}
}

func TestCanteenReviewConcurrentVersionConflict(t *testing.T) {
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
	hall := newCanteenHall(t, halls)
	store := newCanteenStore(t, stores, hall.ID, "并发测试店", canteen.StoreStatusActive)
	review := &canteen.ReviewEvent{
		StoreID: store.ID, UserID: 42, VisitedAt: time.Now().Add(-time.Hour),
		Dimensions: canteen.Dimensions{Taste: 5, Value: 5, Queue: 5, Hygiene: 5, Service: 5},
	}
	review.ApplyCreate(canteen.DefaultStoreWeights)
	if err := reviews.Create(ctx, review); err != nil {
		t.Fatalf("create review: %v", err)
	}
	expected := review.UpdatedAt

	type update struct {
		review   *canteen.ReviewEvent
		revision canteen.Revision
	}
	updates := make([]update, 2)
	for i, content := range []string{"并发修改 A", "并发修改 B"} {
		copied := *review
		revision := copied.MakeRevision(42, time.Now())
		copied.ApplyUpdate(canteen.UpdateReview{
			Dimensions: canteen.Dimensions{Taste: 4, Value: 4, Queue: 4, Hygiene: 4, Service: 4},
			Content:    content,
			Now:        time.Now(),
		}, canteen.DefaultStoreWeights)
		updates[i] = update{review: &copied, revision: revision}
	}

	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, item := range updates {
		wg.Add(1)
		go func(item update) {
			defer wg.Done()
			results <- reviews.Update(ctx, item.review, item.revision, expected)
		}(item)
	}
	wg.Wait()
	close(results)

	success, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, canteen.ErrReviewConflict):
			conflicts++
		default:
			t.Fatalf("unexpected update error: %v", err)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflicts=%d, want 1/1", success, conflicts)
	}
}
