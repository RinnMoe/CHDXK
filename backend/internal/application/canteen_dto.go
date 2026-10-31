package application

import (
	"fmt"
	"time"

	"jcourse/internal/domain/canteen"
)

// CanteenHallDTO is the hall directory row.
type CanteenHallDTO struct {
	ID           int    `json:"id"`
	CampusKey    string `json:"campus_key"`
	Name         string `json:"name"`
	ShortName    string `json:"short_name"`
	Address      string `json:"address"`
	Status       string `json:"status"`
	DisplayOrder int    `json:"display_order"`
	StoreCount   int    `json:"store_count"`
}

// CanteenRatingDTO is the persisted store rating summary.
type CanteenRatingDTO struct {
	Count int     `json:"count"`
	Avg   float64 `json:"avg"`
	Score float64 `json:"score"`
}

// CanteenDimensionsDTO is the five-dimension average.
type CanteenDimensionsDTO struct {
	Taste   float64 `json:"taste"`
	Value   float64 `json:"value"`
	Queue   float64 `json:"queue"`
	Hygiene float64 `json:"hygiene"`
	Service float64 `json:"service"`
}

// CanteenStoreListItemDTO is the store list/ranking card.
type CanteenStoreListItemDTO struct {
	ID           int              `json:"id"`
	HallID       int              `json:"hall_id"`
	HallName     string           `json:"hall_name"`
	CampusKey    string           `json:"campus_key"`
	Name         string           `json:"name"`
	Floor        string           `json:"floor"`
	Categories   []string         `json:"categories"`
	Status       string           `json:"status"`
	CoverURL     string           `json:"cover_url"`
	Rating       CanteenRatingDTO `json:"rating"`
	ReviewCount  int              `json:"review_count"`
	LastReviewAt *time.Time       `json:"last_review_at"`
	// Ranking explanation fields; empty on ordinary directory responses.
	Confidence     string   `json:"confidence,omitempty"`
	SummaryTags    []string `json:"summary_tags"`
	DishCount      int      `json:"dish_count"`
	DishPhotoCount int      `json:"dish_photo_count"`
}

// CanteenMySummaryDTO is the viewer's own weighted sample for a store.
type CanteenMySummaryDTO struct {
	WeightedScore      float64    `json:"weighted_score"`
	WindowSize         int        `json:"window_size"`
	LifetimeEventCount int        `json:"lifetime_event_count"`
	LastVisitedAt      *time.Time `json:"last_visited_at"`
}

// CanteenMediaDTO is the public image projection. Pending images are never
// returned without their own review context.
type CanteenMediaDTO struct {
	ID     int    `json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// CanteenReviewDTO is the review projection.
type CanteenReviewDTO struct {
	ID           int               `json:"id"`
	StoreID      int               `json:"store_id"`
	StoreName    string            `json:"store_name"`
	UserID       int               `json:"user_id"`
	VisitedAt    time.Time         `json:"visited_at"`
	Taste        int               `json:"taste"`
	Value        int               `json:"value"`
	Queue        int               `json:"queue"`
	Hygiene      int               `json:"hygiene"`
	Service      int               `json:"service"`
	OverallScore float64           `json:"overall_score"`
	Content      string            `json:"content"`
	Tags         []string          `json:"tags"`
	Status       string            `json:"status"`
	LikeCount    int               `json:"like_count"`
	DislikeCount int               `json:"dislike_count"`
	MyVote       int               `json:"my_vote"`
	Media        []CanteenMediaDTO `json:"media"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// CanteenDishDTO is the dish projection.
type CanteenDishDTO struct {
	ID          int     `json:"id"`
	StoreID     int     `json:"store_id"`
	StoreName   string  `json:"store_name"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	RatingCount int     `json:"rating_count"`
	RatingAvg   float64 `json:"rating_avg"`
	RatingScore float64 `json:"rating_score"`
	TasteAvg    float64 `json:"taste_avg"`
	PortionAvg  float64 `json:"portion_avg"`
	ValueAvg    float64 `json:"value_avg"`
	ReviewCount int     `json:"review_count"`
	CoverURL    string  `json:"cover_url"`
	PhotoCount  int     `json:"photo_count"`
}

// CanteenStoreDetailDTO is the store detail page payload.
type CanteenStoreDetailDTO struct {
	CanteenStoreListItemDTO
	Dimensions     CanteenDimensionsDTO `json:"dimensions"`
	Dishes         []CanteenDishDTO     `json:"dishes"`
	MySummary      *CanteenMySummaryDTO `json:"my_summary"`
	MyLatestReview *CanteenReviewDTO    `json:"my_latest_review"`
}

// CanteenHomeStatsDTO summarises the domain for the discovery page.
type CanteenHomeStatsDTO struct {
	StoreCount  int `json:"store_count"`
	ReviewCount int `json:"review_count"`
	DishCount   int `json:"dish_count"`
}

// CanteenHomeDTO is the discovery aggregate.
type CanteenHomeDTO struct {
	Halls         []CanteenHallDTO          `json:"halls"`
	TopStores     []CanteenStoreListItemDTO `json:"top_stores"`
	NewStores     []CanteenStoreListItemDTO `json:"new_stores"`
	LatestReviews []CanteenReviewDTO        `json:"latest_reviews"`
	HotDishes     []CanteenDishDTO          `json:"hot_dishes"`
	Hero          *CanteenHomeHeroDTO       `json:"hero"`
	Feeds         []CanteenHomeFeedDTO      `json:"feeds"`
	Stats         CanteenHomeStatsDTO       `json:"stats"`
}

// CanteenHomeHeroDTO is the "what to eat today" card.
type CanteenHomeHeroDTO struct {
	Type        string  `json:"type"`
	StoreID     int     `json:"store_id"`
	StoreName   string  `json:"store_name"`
	HallName    string  `json:"hall_name"`
	DishID      int     `json:"dish_id"`
	DishName    string  `json:"dish_name"`
	ImageURL    string  `json:"image_url"`
	Reason      string  `json:"reason"`
	RatingScore float64 `json:"rating_score"`
	RatingCount int     `json:"rating_count"`
}

// CanteenHomeFeedDTO is one decision-oriented discovery card.
type CanteenHomeFeedDTO struct {
	Type        string   `json:"type"`
	StoreID     int      `json:"store_id"`
	StoreName   string   `json:"store_name"`
	HallName    string   `json:"hall_name"`
	DishID      int      `json:"dish_id"`
	DishName    string   `json:"dish_name"`
	Reason      string   `json:"reason"`
	ImageURL    string   `json:"image_url"`
	RatingScore float64  `json:"rating_score"`
	RatingCount int      `json:"rating_count"`
	Tags        []string `json:"tags"`
}

// CanteenSearchDTO is the unified canteen + dish search result.
type CanteenSearchDTO struct {
	Stores []CanteenStoreListItemDTO `json:"stores"`
	Dishes []CanteenDishDTO          `json:"dishes"`
}

// CanteenDishDetailDTO is the dish detail page payload.
type CanteenDishDetailDTO struct {
	Dish        CanteenDishDTO         `json:"dish"`
	Photos      []CanteenMediaDTO      `json:"photos"`
	Reviews     []CanteenDishReviewDTO `json:"reviews"`
	ReviewTotal int64                  `json:"review_total"`
}

// CanteenStorePage is the paginated store result.
type CanteenStorePage struct {
	Items    []CanteenStoreListItemDTO `json:"items"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

// CanteenReviewPage is the paginated review result.
type CanteenReviewPage struct {
	Items      []CanteenReviewDTO `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

// CanteenDishReviewDTO is a user's own dish rating.
type CanteenDishReviewDTO struct {
	ID           int       `json:"id"`
	DishID       int       `json:"dish_id"`
	DishName     string    `json:"dish_name"`
	StoreID      int       `json:"store_id"`
	StoreName    string    `json:"store_name"`
	Taste        int       `json:"taste"`
	Portion      int       `json:"portion"`
	Value        int       `json:"value"`
	OverallScore float64   `json:"overall_score"`
	Content      string    `json:"content"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// CanteenMediaAssetDTO is a media row including moderation status.
type CanteenMediaAssetDTO struct {
	ID        int       `json:"id"`
	URL       string    `json:"url"`
	MIME      string    `json:"mime"`
	ByteSize  int       `json:"byte_size"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// CanteenAdminMediaDTO adds ownership information for the moderation queue.
type CanteenAdminMediaDTO struct {
	CanteenMediaAssetDTO
	OwnerUserID int `json:"owner_user_id"`
}

// CanteenReportDTO is the report queue projection.
type CanteenReportDTO struct {
	ID             int        `json:"id"`
	ReporterUserID int        `json:"reporter_user_id"`
	SubjectType    string     `json:"subject_type"`
	SubjectID      int        `json:"subject_id"`
	ReasonCode     string     `json:"reason_code"`
	Detail         string     `json:"detail"`
	Status         string     `json:"status"`
	HandledBy      *int       `json:"handled_by"`
	HandledAt      *time.Time `json:"handled_at"`
	ResolutionNote string     `json:"resolution_note"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CanteenContributionsDTO bundles the caller's dish ratings and uploads.
type CanteenContributionsDTO struct {
	Dishes []CanteenDishReviewDTO `json:"dishes"`
	Media  []CanteenMediaAssetDTO `json:"media"`
}

func newCanteenDishReviewDTO(view canteen.DishReviewView) CanteenDishReviewDTO {
	return CanteenDishReviewDTO{
		ID:           view.ID,
		DishID:       view.DishID,
		DishName:     view.DishName,
		StoreID:      view.StoreID,
		StoreName:    view.StoreName,
		Taste:        view.Taste,
		Portion:      view.Portion,
		Value:        view.Value,
		OverallScore: view.OverallScore,
		Content:      view.Content,
		Status:       string(view.Status),
		CreatedAt:    view.CreatedAt,
	}
}

func newCanteenMediaAssetDTO(asset canteen.Media) CanteenMediaAssetDTO {
	return CanteenMediaAssetDTO{
		ID:        asset.ID,
		URL:       canteenMediaURL(asset.ID),
		MIME:      asset.MIME,
		ByteSize:  asset.ByteSize,
		Width:     asset.Width,
		Height:    asset.Height,
		Status:    string(asset.Status),
		CreatedAt: asset.CreatedAt,
	}
}

func newCanteenAdminMediaDTO(asset canteen.Media) CanteenAdminMediaDTO {
	return CanteenAdminMediaDTO{
		CanteenMediaAssetDTO: newCanteenMediaAssetDTO(asset),
		OwnerUserID:          asset.OwnerUserID,
	}
}

func newCanteenReportDTO(report canteen.Report) CanteenReportDTO {
	return CanteenReportDTO{
		ID:             report.ID,
		ReporterUserID: report.ReporterUserID,
		SubjectType:    string(report.SubjectType),
		SubjectID:      report.SubjectID,
		ReasonCode:     string(report.ReasonCode),
		Detail:         report.Detail,
		Status:         string(report.Status),
		HandledBy:      report.HandledBy,
		HandledAt:      report.HandledAt,
		ResolutionNote: report.ResolutionNote,
		CreatedAt:      report.CreatedAt,
	}
}

func canteenMediaURL(mediaID int) string {
	if mediaID <= 0 {
		return ""
	}
	return fmt.Sprintf("/api/canteen/media/%d", mediaID)
}

func canteenCoverURL(mediaID *int) string {
	if mediaID == nil || *mediaID == 0 {
		return ""
	}
	return fmt.Sprintf("/api/canteen/media/%d", *mediaID)
}

func newCanteenHallDTO(view canteen.HallView) CanteenHallDTO {
	return CanteenHallDTO{
		ID:           view.ID,
		CampusKey:    view.CampusKey,
		Name:         view.Name,
		ShortName:    view.ShortName,
		Address:      view.Address,
		Status:       string(view.Status),
		DisplayOrder: view.DisplayOrder,
		StoreCount:   view.StoreCount,
	}
}

func newCanteenStoreListItemDTO(item canteen.StoreListItem) CanteenStoreListItemDTO {
	return CanteenStoreListItemDTO{
		ID:             item.ID,
		HallID:         item.HallID,
		HallName:       item.HallName,
		CampusKey:      item.CampusKey,
		Name:           item.Name,
		Floor:          item.Floor,
		Categories:     append([]string{}, item.Categories...),
		Status:         string(item.Status),
		CoverURL:       canteenCoverURL(item.CoverMediaID),
		Rating:         CanteenRatingDTO{Count: item.RatingCount, Avg: item.RatingAvg, Score: item.RatingScore},
		ReviewCount:    item.ReviewEventCount,
		LastReviewAt:   item.LastReviewAt,
		Confidence:     item.Confidence,
		SummaryTags:    append([]string{}, item.SummaryTags...),
		DishCount:      item.DishCount,
		DishPhotoCount: item.DishPhotoCount,
	}
}

func newCanteenReviewDTO(view canteen.ReviewView) CanteenReviewDTO {
	media := make([]CanteenMediaDTO, 0, len(view.Media))
	for _, item := range view.Media {
		media = append(media, CanteenMediaDTO{ID: item.ID, URL: item.URL, Width: item.Width, Height: item.Height})
	}
	return CanteenReviewDTO{
		ID:           view.ID,
		StoreID:      view.StoreID,
		StoreName:    view.StoreName,
		UserID:       view.UserID,
		VisitedAt:    view.VisitedAt,
		Taste:        view.Dimensions.Taste,
		Value:        view.Dimensions.Value,
		Queue:        view.Dimensions.Queue,
		Hygiene:      view.Dimensions.Hygiene,
		Service:      view.Dimensions.Service,
		OverallScore: view.OverallScore,
		Content:      view.Content,
		Tags:         append([]string{}, view.Tags...),
		Status:       string(view.Status),
		LikeCount:    view.LikeCount,
		DislikeCount: view.DislikeCount,
		MyVote:       view.MyVote,
		Media:        media,
		CreatedAt:    view.CreatedAt,
		UpdatedAt:    view.UpdatedAt,
	}
}

func newCanteenDishDTO(view canteen.DishView) CanteenDishDTO {
	return CanteenDishDTO{
		ID:          view.ID,
		StoreID:     view.StoreID,
		StoreName:   view.StoreName,
		Name:        view.Name,
		Status:      string(view.Status),
		RatingCount: view.RatingCount,
		RatingAvg:   view.RatingAvg,
		RatingScore: view.RatingScore,
		TasteAvg:    view.TasteAvg,
		PortionAvg:  view.PortionAvg,
		ValueAvg:    view.ValueAvg,
		ReviewCount: view.ReviewEventCount,
		CoverURL:    view.CoverURL,
		PhotoCount:  view.PhotoCount,
	}
}
