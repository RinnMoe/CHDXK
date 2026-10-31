package canteen

import (
	"context"
	"time"
)

// HallQuery reads the hall directory.
type HallQuery interface {
	ListHalls(ctx context.Context) ([]HallView, error)
}

// StoreQuery reads stores, rankings and the discovery aggregate.
type StoreQuery interface {
	ListStores(ctx context.Context, filter StoreListFilter) ([]StoreListItem, int64, error)
	GetStoreDetail(ctx context.Context, id int, viewerID int) (*StoreDetail, error)
	ListRankings(ctx context.Context, filter RankingFilter) ([]StoreListItem, int64, error)
	GetHome(ctx context.Context) (*HomeView, error)
}

// ReviewQuery reads visit reviews and the caller's own aggregate state.
type ReviewQuery interface {
	ListReviews(ctx context.Context, filter ReviewListFilter) ([]ReviewView, int64, error)
	GetReview(ctx context.Context, id int) (*ReviewView, error)
	ListUserReviews(ctx context.Context, userID int, page, pageSize int) ([]ReviewView, int64, error)
	GetUserSummary(ctx context.Context, storeID, userID int) (*UserRatingSummary, error)
	CountSince(ctx context.Context, userID int, since time.Time) (int, error)
	ListRecentContent(ctx context.Context, userID int, since time.Time, limit int) ([]string, error)
}

// DishQuery reads the dish directory and a user's dish ratings.
type DishQuery interface {
	ListDishes(ctx context.Context, filter DishListFilter) ([]DishView, error)
	GetDish(ctx context.Context, id int) (*Dish, error)
	ListDishPhotos(ctx context.Context, dishID int) ([]MediaView, error)
	ListDishReviews(ctx context.Context, dishID int, page, pageSize int) ([]DishReviewView, int64, error)
	ListDishReviewsByUser(ctx context.Context, userID int, page, pageSize int) ([]DishReviewView, int64, error)
}

// MediaQuery reads media metadata. Bytes are served through MediaStorage.
type MediaQuery interface {
	GetMedia(ctx context.Context, id int) (*Media, error)
	ListMediaByOwner(ctx context.Context, userID int, page, pageSize int) ([]Media, int64, error)
	ListMediaByStatus(ctx context.Context, filter MediaFilter) ([]Media, int64, error)
	// IsMediaPublic reports whether at least one live parent (active review or
	// active dish) currently exposes this approved media.
	IsMediaPublic(ctx context.Context, id int) (bool, error)
}

// DimensionAverages is the read-side average of the five store dimensions.
type DimensionAverages struct {
	Taste   float64
	Value   float64
	Queue   float64
	Hygiene float64
	Service float64
}

// HallView is the read model for the hall directory.
type HallView struct {
	ID           int
	CampusKey    string
	Name         string
	ShortName    string
	Address      string
	Status       HallStatus
	DisplayOrder int
	StoreCount   int
}

// StoreListItem is the list/ranking read model. Aggregates come from the
// persisted stats columns, so list queries never scan review events.
type StoreListItem struct {
	ID               int
	HallID           int
	HallName         string
	CampusKey        string
	Name             string
	Floor            string
	Categories       []string
	Status           StoreStatus
	CoverMediaID     *int
	RatingCount      int
	RatingAvg        float64
	RatingScore      float64
	ReviewEventCount int
	LastReviewAt     *time.Time
	// Ranking explanation fields. They are populated for ranking responses and
	// left empty for ordinary directory lists.
	Confidence     string
	SummaryTags    []string
	DishCount      int
	DishPhotoCount int
}

// PubliclyVisible reports whether the store may be shown outside admin views.
// Offline stores keep their history readable.
func (s StoreListItem) PubliclyVisible() bool {
	return s.Status == StoreStatusActive || s.Status == StoreStatusOffline
}

// StoreDetail is the single-store read model.
type StoreDetail struct {
	StoreListItem
	Store          *Store
	Dimensions     DimensionAverages
	Dishes         []DishView
	MySummary      *UserRatingSummary
	MyLatestReview *ReviewView
}

// UserRatingSummary is one user's weighted sample for a store.
type UserRatingSummary struct {
	StoreID            int
	UserID             int
	WeightedScore      float64
	WindowSize         int
	LifetimeEventCount int
	LastVisitedAt      *time.Time
	ComputedAt         time.Time
}

// MediaView is the public projection of an uploaded image.
type MediaView struct {
	ID        int
	Status    MediaStatus
	Width     int
	Height    int
	URL       string
	SortOrder int
}

// ReviewView is the review read model with author, media and vote state.
type ReviewView struct {
	ID         int
	StoreID    int
	StoreName  string
	UserID     int
	AuthorName string
	VisitedAt  time.Time
	Dimensions
	OverallScore float64
	Content      string
	Tags         []string
	Status       ReviewStatus
	LikeCount    int
	DislikeCount int
	MyVote       int
	Media        []MediaView
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DishView is the dish read model.
type DishView struct {
	ID               int
	StoreID          int
	StoreName        string
	Name             string
	Status           DishStatus
	RatingCount      int
	RatingAvg        float64
	RatingScore      float64
	TasteAvg         float64
	PortionAvg       float64
	ValueAvg         float64
	ReviewEventCount int
	CoverURL         string
	PhotoCount       int
}

// HomeStats summarises the canteen domain for the discovery page.
type HomeStats struct {
	StoreCount  int
	ReviewCount int
	DishCount   int
}

// HomeHero is the "what to eat today" recommendation.
type HomeHero struct {
	Type        string
	StoreID     int
	StoreName   string
	HallName    string
	DishID      int
	DishName    string
	ImageURL    string
	Reason      string
	RatingScore float64
	RatingCount int
}

// HomeFeedItem is one decision-oriented discovery card.
type HomeFeedItem struct {
	Type        string
	StoreID     int
	StoreName   string
	HallName    string
	DishID      int
	DishName    string
	Reason      string
	ImageURL    string
	RatingScore float64
	RatingCount int
	Tags        []string
}

// HomeView is the discovery aggregate. Empty slices mean a real cold start, not
// an error, and the frontend renders an explicit empty state.
type HomeView struct {
	Halls         []HallView
	TopStores     []StoreListItem
	NewStores     []StoreListItem
	LatestReviews []ReviewView
	HotDishes     []DishView
	Hero          *HomeHero
	Feeds         []HomeFeedItem
	Stats         HomeStats
}

// SearchView is the unified canteen + dish search result.
type SearchView struct {
	Stores []StoreListItem
	Dishes []DishView
}

// StoreListFilter drives store search and discovery.
type StoreListFilter struct {
	Q         string
	CampusKey string
	HallID    int
	Category  string
	Status    StoreStatus
	Sort      string
	Page      int
	PageSize  int
}

// RankingFilter drives the cached ranking page.
type RankingFilter struct {
	CampusKey string
	HallID    int
	Sort      string
	Page      int
	PageSize  int
}

// DishListFilter drives dish lists and suggestions.
type DishListFilter struct {
	StoreID int
	Status  DishStatus
	Q       string
	Limit   int
}
