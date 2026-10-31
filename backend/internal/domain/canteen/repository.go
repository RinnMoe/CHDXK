package canteen

import (
	"context"
	"time"
)

// HallRepository persists canteen buildings.
type HallRepository interface {
	Create(ctx context.Context, hall *Hall) error
	Update(ctx context.Context, hall *Hall) error
	Get(ctx context.Context, id int) (*Hall, error)
}

// StoreRepository persists stores and their persisted aggregate stats.
type StoreRepository interface {
	Create(ctx context.Context, store *Store) error
	Update(ctx context.Context, store *Store) error
	Get(ctx context.Context, id int) (*Store, error)
	GetByHallAndNormalizedName(ctx context.Context, hallID int, normalizedName string) (*Store, error)
	UpdateStatus(ctx context.Context, id int, status StoreStatus, now time.Time) error
	Merge(ctx context.Context, id, targetID int, now time.Time) error
	UpdateStats(ctx context.Context, id int, stats StoreStats, now time.Time) error
}

// DishRepository persists dishes and their persisted aggregate stats.
type DishRepository interface {
	Create(ctx context.Context, dish *Dish) error
	Update(ctx context.Context, dish *Dish) error
	Get(ctx context.Context, id int) (*Dish, error)
	GetByStoreAndNormalizedName(ctx context.Context, storeID int, normalizedName string) (*Dish, error)
	UpdateStatus(ctx context.Context, id int, status DishStatus, now time.Time) error
	Merge(ctx context.Context, id, targetID int, now time.Time) error
	UpdateStats(ctx context.Context, id int, stats DishStats, now time.Time) error
}

// ReviewRepository persists visit events and revision snapshots. Create,
// Update, SoftDelete and Hide recompute the affected user summaries and store
// stats inside the same transaction.
type ReviewRepository interface {
	Create(ctx context.Context, review *ReviewEvent) error
	Update(ctx context.Context, review *ReviewEvent, revision Revision, expectedUpdatedAt time.Time) error
	Get(ctx context.Context, id int) (*ReviewEvent, error)
	FindByClientRequestID(ctx context.Context, userID int, requestID string) (*ReviewEvent, error)
	SoftDelete(ctx context.Context, id int, now time.Time) error
	Hide(ctx context.Context, id int, byUserID int, reason string, now time.Time) error
	RefreshStoreStats(ctx context.Context, storeID int) error
	RefreshAllStoreStats(ctx context.Context) error
}

// VoteRepository persists helpfulness votes. One row per user and review.
type VoteRepository interface {
	Get(ctx context.Context, reviewEventID, userID int) (*Vote, error)
	Save(ctx context.Context, vote *Vote) error
	Delete(ctx context.Context, reviewEventID, userID int) error
}

// DishReviewRepository persists dish ratings and recomputes dish aggregates in
// the same transaction.
type DishReviewRepository interface {
	Create(ctx context.Context, review *DishReviewEvent) error
}

// MediaRepository persists media metadata. Bytes live behind MediaStorage.
type MediaRepository interface {
	Create(ctx context.Context, media *Media) error
	Update(ctx context.Context, media *Media) error
	ListByReview(ctx context.Context, reviewID int) ([]Media, error)
	AttachToReview(ctx context.Context, reviewID, mediaID int, sortOrder int) error
	AttachToDish(ctx context.Context, dishID, mediaID, userID int, sortOrder int, now time.Time) error
	CountByOwnerSince(ctx context.Context, userID int, since time.Time) (int, error)
}

// ReportRepository persists user reports and admin resolutions.
type ReportRepository interface {
	Create(ctx context.Context, report *Report) error
	Get(ctx context.Context, id int) (*Report, error)
	List(ctx context.Context, filter ReportFilter) ([]Report, int64, error)
	Resolve(ctx context.Context, id int, handlerID int, status ReportStatus, note string, now time.Time) error
}
