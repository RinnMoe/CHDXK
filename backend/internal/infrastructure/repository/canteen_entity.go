package repository

import (
	"time"

	"github.com/lib/pq"

	"jcourse/internal/domain/canteen"
)

func canteenStringArray(values []string) pq.StringArray {
	if len(values) == 0 {
		return pq.StringArray{}
	}
	return pq.StringArray(values)
}

type CanteenHallEntity struct {
	ID              int       `gorm:"column:id"`
	CampusKey       string    `gorm:"column:campus_key;uniqueIndex:uniq_canteen_halls_campus_name"`
	Name            string    `gorm:"column:name;uniqueIndex:uniq_canteen_halls_campus_name"`
	ShortName       string    `gorm:"column:short_name"`
	Address         string    `gorm:"column:address"`
	Latitude        *float64  `gorm:"column:latitude;type:numeric(9,6)"`
	Longitude       *float64  `gorm:"column:longitude;type:numeric(9,6)"`
	ExternalPlaceID string    `gorm:"column:external_place_id"`
	Status          string    `gorm:"column:status"`
	DisplayOrder    int       `gorm:"column:display_order"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (CanteenHallEntity) TableName() string {
	return "canteen_halls"
}

type CanteenStoreEntity struct {
	ID                int            `gorm:"column:id"`
	HallID            int            `gorm:"column:hall_id;uniqueIndex:uniq_canteen_stores_hall_name,priority:1"`
	Name              string         `gorm:"column:name"`
	NormalizedName    string         `gorm:"column:normalized_name;uniqueIndex:uniq_canteen_stores_hall_name,priority:2,where:status <> 'merged'"`
	Floor             string         `gorm:"column:floor"`
	Categories        pq.StringArray `gorm:"column:categories;type:text[]"`
	Status            string         `gorm:"column:status"`
	MergedIntoStoreID *int           `gorm:"column:merged_into_store_id"`
	CoverMediaID      *int           `gorm:"column:cover_media_id"`
	RatingCount       int            `gorm:"column:rating_count"`
	RatingAvg         float64        `gorm:"column:rating_avg"`
	RatingScore       float64        `gorm:"column:rating_score"`
	TasteAvg          float64        `gorm:"column:taste_avg"`
	ValueAvg          float64        `gorm:"column:value_avg"`
	QueueAvg          float64        `gorm:"column:queue_avg"`
	HygieneAvg        float64        `gorm:"column:hygiene_avg"`
	ServiceAvg        float64        `gorm:"column:service_avg"`
	ReviewEventCount  int            `gorm:"column:review_event_count"`
	LastReviewAt      *time.Time     `gorm:"column:last_review_at"`
	RatingUpdatedAt   *time.Time     `gorm:"column:rating_updated_at"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at"`
}

func (CanteenStoreEntity) TableName() string {
	return "canteen_stores"
}

type CanteenDishEntity struct {
	ID               int        `gorm:"column:id"`
	StoreID          int        `gorm:"column:store_id;uniqueIndex:uniq_canteen_dishes_store_name,priority:1"`
	Name             string     `gorm:"column:name"`
	NormalizedName   string     `gorm:"column:normalized_name;uniqueIndex:uniq_canteen_dishes_store_name,priority:2,where:status <> 'merged'"`
	Status           string     `gorm:"column:status"`
	MergedIntoDishID *int       `gorm:"column:merged_into_dish_id"`
	RatingCount      int        `gorm:"column:rating_count"`
	RatingAvg        float64    `gorm:"column:rating_avg"`
	RatingScore      float64    `gorm:"column:rating_score"`
	TasteAvg         float64    `gorm:"column:taste_avg"`
	PortionAvg       float64    `gorm:"column:portion_avg"`
	ValueAvg         float64    `gorm:"column:value_avg"`
	ReviewEventCount int        `gorm:"column:review_event_count"`
	RatingUpdatedAt  *time.Time `gorm:"column:rating_updated_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (CanteenDishEntity) TableName() string {
	return "canteen_dishes"
}

type CanteenDishAliasEntity struct {
	ID             int       `gorm:"column:id"`
	DishID         int       `gorm:"column:dish_id;uniqueIndex:uniq_canteen_dish_aliases_dish_name"`
	Name           string    `gorm:"column:name"`
	NormalizedName string    `gorm:"column:normalized_name;uniqueIndex:uniq_canteen_dish_aliases_dish_name"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (CanteenDishAliasEntity) TableName() string {
	return "canteen_dish_aliases"
}

// CanteenDishMediaEntity links an approved/pending media asset to a dish. The
// public dish gallery only exposes media whose own status is approved.
type CanteenDishMediaEntity struct {
	ID        int       `gorm:"column:id"`
	DishID    int       `gorm:"column:dish_id;uniqueIndex:uniq_canteen_dish_media"`
	MediaID   int       `gorm:"column:media_id;uniqueIndex:uniq_canteen_dish_media"`
	UserID    int       `gorm:"column:user_id"`
	SortOrder int       `gorm:"column:sort_order"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (CanteenDishMediaEntity) TableName() string {
	return "canteen_dish_media"
}

type CanteenReviewEventEntity struct {
	ID                  int            `gorm:"column:id"`
	StoreID             int            `gorm:"column:store_id"`
	UserID              int            `gorm:"column:user_id;uniqueIndex:uniq_canteen_review_events_user_request,priority:1,where:client_request_id IS NOT NULL"`
	VisitedAt           time.Time      `gorm:"column:visited_at"`
	Taste               int            `gorm:"column:taste"`
	Value               int            `gorm:"column:value"`
	Queue               int            `gorm:"column:queue"`
	Hygiene             int            `gorm:"column:hygiene"`
	Service             int            `gorm:"column:service"`
	OverallScore        float64        `gorm:"column:overall_score;type:numeric(3,2)"`
	ScoreFormulaVersion int            `gorm:"column:score_formula_version"`
	Content             string         `gorm:"column:content"`
	Tags                pq.StringArray `gorm:"column:tags;type:text[]"`
	Status              string         `gorm:"column:status"`
	LikeCount           int            `gorm:"column:like_count"`
	DislikeCount        int            `gorm:"column:dislike_count"`
	ClientRequestID     *string        `gorm:"column:client_request_id;type:uuid;uniqueIndex:uniq_canteen_review_events_user_request,priority:2,where:client_request_id IS NOT NULL"`
	HiddenReason        string         `gorm:"column:hidden_reason"`
	HiddenBy            *int           `gorm:"column:hidden_by"`
	HiddenAt            *time.Time     `gorm:"column:hidden_at"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at"`
}

func (CanteenReviewEventEntity) TableName() string {
	return "canteen_review_events"
}

type CanteenReviewRevisionEntity struct {
	ID            int            `gorm:"column:id"`
	ReviewEventID int            `gorm:"column:review_event_id"`
	CreatedBy     int            `gorm:"column:created_by"`
	VisitedAt     time.Time      `gorm:"column:visited_at"`
	Taste         int            `gorm:"column:taste"`
	Value         int            `gorm:"column:value"`
	Queue         int            `gorm:"column:queue"`
	Hygiene       int            `gorm:"column:hygiene"`
	Service       int            `gorm:"column:service"`
	OverallScore  float64        `gorm:"column:overall_score;type:numeric(3,2)"`
	Content       string         `gorm:"column:content"`
	Tags          pq.StringArray `gorm:"column:tags;type:text[]"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
}

func (CanteenReviewRevisionEntity) TableName() string {
	return "canteen_review_revisions"
}

type CanteenUserRatingSummaryEntity struct {
	StoreID            int        `gorm:"column:store_id;primaryKey"`
	UserID             int        `gorm:"column:user_id;primaryKey"`
	WeightedScore      float64    `gorm:"column:weighted_score;type:numeric(3,2)"`
	WindowSize         int        `gorm:"column:window_size"`
	LifetimeEventCount int        `gorm:"column:lifetime_event_count"`
	LastVisitedAt      *time.Time `gorm:"column:last_visited_at"`
	ComputedAt         time.Time  `gorm:"column:computed_at"`
}

func (CanteenUserRatingSummaryEntity) TableName() string {
	return "canteen_user_rating_summaries"
}

type CanteenReviewVoteEntity struct {
	ReviewEventID int       `gorm:"column:review_event_id;primaryKey"`
	UserID        int       `gorm:"column:user_id;primaryKey"`
	VoteType      int       `gorm:"column:vote_type"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (CanteenReviewVoteEntity) TableName() string {
	return "canteen_review_votes"
}

type CanteenReviewDishEntity struct {
	ReviewEventID int       `gorm:"column:review_event_id;primaryKey"`
	DishID        int       `gorm:"column:dish_id"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (CanteenReviewDishEntity) TableName() string {
	return "canteen_review_dishes"
}

type CanteenDishReviewEventEntity struct {
	ID            int       `gorm:"column:id"`
	DishID        int       `gorm:"column:dish_id"`
	UserID        int       `gorm:"column:user_id"`
	ReviewEventID *int      `gorm:"column:review_event_id"`
	Taste         int       `gorm:"column:taste"`
	Portion       int       `gorm:"column:portion"`
	Value         int       `gorm:"column:value"`
	OverallScore  float64   `gorm:"column:overall_score;type:numeric(3,2)"`
	Content       string    `gorm:"column:content"`
	Status        string    `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (CanteenDishReviewEventEntity) TableName() string {
	return "canteen_dish_review_events"
}

type CanteenDishRatingSummaryEntity struct {
	DishID             int        `gorm:"column:dish_id;primaryKey"`
	UserID             int        `gorm:"column:user_id;primaryKey"`
	WeightedScore      float64    `gorm:"column:weighted_score;type:numeric(3,2)"`
	WindowSize         int        `gorm:"column:window_size"`
	LifetimeEventCount int        `gorm:"column:lifetime_event_count"`
	LastReviewedAt     *time.Time `gorm:"column:last_reviewed_at"`
	ComputedAt         time.Time  `gorm:"column:computed_at"`
}

func (CanteenDishRatingSummaryEntity) TableName() string {
	return "canteen_dish_rating_summaries"
}

type CanteenMediaEntity struct {
	ID            int        `gorm:"column:id"`
	OwnerUserID   int        `gorm:"column:owner_user_id"`
	Purpose       string     `gorm:"column:purpose"`
	StorageDriver string     `gorm:"column:storage_driver"`
	StorageKey    string     `gorm:"column:storage_key;uniqueIndex:uniq_canteen_media_storage_key"`
	MIME          string     `gorm:"column:mime"`
	ByteSize      int        `gorm:"column:byte_size"`
	Width         int        `gorm:"column:width"`
	Height        int        `gorm:"column:height"`
	SHA256        string     `gorm:"column:sha256"`
	Status        string     `gorm:"column:status"`
	ReviewedBy    *int       `gorm:"column:reviewed_by"`
	ReviewedAt    *time.Time `gorm:"column:reviewed_at"`
	RejectReason  string     `gorm:"column:reject_reason"`
	ExpiresAt     *time.Time `gorm:"column:expires_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (CanteenMediaEntity) TableName() string {
	return "canteen_media"
}

type CanteenReviewMediaEntity struct {
	ReviewEventID int `gorm:"column:review_event_id;primaryKey"`
	MediaID       int `gorm:"column:media_id;primaryKey;uniqueIndex:uniq_canteen_review_media_media"`
	SortOrder     int `gorm:"column:sort_order"`
}

func (CanteenReviewMediaEntity) TableName() string {
	return "canteen_review_media"
}

type CanteenReportEntity struct {
	ID             int        `gorm:"column:id"`
	ReporterUserID int        `gorm:"column:reporter_user_id;uniqueIndex:uniq_canteen_reports_open_subject,priority:1,where:status <> 'resolved' AND status <> 'dismissed'"`
	SubjectType    string     `gorm:"column:subject_type;uniqueIndex:uniq_canteen_reports_open_subject,priority:2,where:status <> 'resolved' AND status <> 'dismissed'"`
	SubjectID      int        `gorm:"column:subject_id;uniqueIndex:uniq_canteen_reports_open_subject,priority:3,where:status <> 'resolved' AND status <> 'dismissed'"`
	ReasonCode     string     `gorm:"column:reason_code"`
	Detail         string     `gorm:"column:detail"`
	Status         string     `gorm:"column:status"`
	HandledBy      *int       `gorm:"column:handled_by"`
	HandledAt      *time.Time `gorm:"column:handled_at"`
	ResolutionNote string     `gorm:"column:resolution_note"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (CanteenReportEntity) TableName() string {
	return "canteen_reports"
}

func newCanteenHallEntity(hall *canteen.Hall) CanteenHallEntity {
	return CanteenHallEntity{
		ID:              hall.ID,
		CampusKey:       hall.CampusKey,
		Name:            hall.Name,
		ShortName:       hall.ShortName,
		Address:         hall.Address,
		Latitude:        hall.Latitude,
		Longitude:       hall.Longitude,
		ExternalPlaceID: hall.ExternalPlaceID,
		Status:          string(hall.Status),
		DisplayOrder:    hall.DisplayOrder,
		CreatedAt:       hall.CreatedAt,
		UpdatedAt:       hall.UpdatedAt,
	}
}

func newCanteenHallDomain(entity *CanteenHallEntity) *canteen.Hall {
	if entity == nil {
		return nil
	}
	return &canteen.Hall{
		ID:              entity.ID,
		CampusKey:       entity.CampusKey,
		Name:            entity.Name,
		ShortName:       entity.ShortName,
		Address:         entity.Address,
		Latitude:        entity.Latitude,
		Longitude:       entity.Longitude,
		ExternalPlaceID: entity.ExternalPlaceID,
		Status:          canteen.HallStatus(entity.Status),
		DisplayOrder:    entity.DisplayOrder,
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       entity.UpdatedAt,
	}
}

func newCanteenStoreEntity(store *canteen.Store) CanteenStoreEntity {
	return CanteenStoreEntity{
		ID:                store.ID,
		HallID:            store.HallID,
		Name:              store.Name,
		NormalizedName:    store.NormalizedName,
		Floor:             store.Floor,
		Categories:        canteenStringArray(store.Categories),
		Status:            string(store.Status),
		MergedIntoStoreID: store.MergedIntoStoreID,
		CoverMediaID:      store.CoverMediaID,
		RatingCount:       store.Stats.RatingCount,
		RatingAvg:         store.Stats.RatingAvg,
		RatingScore:       store.Stats.RatingScore,
		TasteAvg:          store.Stats.TasteAvg,
		ValueAvg:          store.Stats.ValueAvg,
		QueueAvg:          store.Stats.QueueAvg,
		HygieneAvg:        store.Stats.HygieneAvg,
		ServiceAvg:        store.Stats.ServiceAvg,
		ReviewEventCount:  store.Stats.ReviewEventCount,
		LastReviewAt:      store.Stats.LastReviewAt,
		RatingUpdatedAt:   store.Stats.RatingUpdatedAt,
		CreatedAt:         store.CreatedAt,
		UpdatedAt:         store.UpdatedAt,
	}
}

func newCanteenStoreDomain(entity *CanteenStoreEntity) *canteen.Store {
	if entity == nil {
		return nil
	}
	return &canteen.Store{
		ID:                entity.ID,
		HallID:            entity.HallID,
		Name:              entity.Name,
		NormalizedName:    entity.NormalizedName,
		Floor:             entity.Floor,
		Categories:        []string(entity.Categories),
		Status:            canteen.StoreStatus(entity.Status),
		MergedIntoStoreID: entity.MergedIntoStoreID,
		CoverMediaID:      entity.CoverMediaID,
		Stats: canteen.StoreStats{
			RatingCount:      entity.RatingCount,
			RatingAvg:        entity.RatingAvg,
			RatingScore:      entity.RatingScore,
			TasteAvg:         entity.TasteAvg,
			ValueAvg:         entity.ValueAvg,
			QueueAvg:         entity.QueueAvg,
			HygieneAvg:       entity.HygieneAvg,
			ServiceAvg:       entity.ServiceAvg,
			ReviewEventCount: entity.ReviewEventCount,
			LastReviewAt:     entity.LastReviewAt,
			RatingUpdatedAt:  entity.RatingUpdatedAt,
		},
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

func newCanteenDishEntity(dish *canteen.Dish) CanteenDishEntity {
	return CanteenDishEntity{
		ID:               dish.ID,
		StoreID:          dish.StoreID,
		Name:             dish.Name,
		NormalizedName:   dish.NormalizedName,
		Status:           string(dish.Status),
		MergedIntoDishID: dish.MergedIntoDishID,
		RatingCount:      dish.Stats.RatingCount,
		RatingAvg:        dish.Stats.RatingAvg,
		RatingScore:      dish.Stats.RatingScore,
		TasteAvg:         dish.Stats.TasteAvg,
		PortionAvg:       dish.Stats.PortionAvg,
		ValueAvg:         dish.Stats.ValueAvg,
		ReviewEventCount: dish.Stats.ReviewEventCount,
		RatingUpdatedAt:  dish.Stats.RatingUpdatedAt,
		CreatedAt:        dish.CreatedAt,
		UpdatedAt:        dish.UpdatedAt,
	}
}

func newCanteenDishDomain(entity *CanteenDishEntity) *canteen.Dish {
	if entity == nil {
		return nil
	}
	return &canteen.Dish{
		ID:               entity.ID,
		StoreID:          entity.StoreID,
		Name:             entity.Name,
		NormalizedName:   entity.NormalizedName,
		Status:           canteen.DishStatus(entity.Status),
		MergedIntoDishID: entity.MergedIntoDishID,
		Stats: canteen.DishStats{
			RatingCount:      entity.RatingCount,
			RatingAvg:        entity.RatingAvg,
			RatingScore:      entity.RatingScore,
			TasteAvg:         entity.TasteAvg,
			PortionAvg:       entity.PortionAvg,
			ValueAvg:         entity.ValueAvg,
			ReviewEventCount: entity.ReviewEventCount,
			RatingUpdatedAt:  entity.RatingUpdatedAt,
		},
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

func newCanteenReviewEntity(review *canteen.ReviewEvent) CanteenReviewEventEntity {
	entity := CanteenReviewEventEntity{
		ID:                  review.ID,
		StoreID:             review.StoreID,
		UserID:              review.UserID,
		VisitedAt:           review.VisitedAt,
		Taste:               review.Dimensions.Taste,
		Value:               review.Dimensions.Value,
		Queue:               review.Dimensions.Queue,
		Hygiene:             review.Dimensions.Hygiene,
		Service:             review.Dimensions.Service,
		OverallScore:        review.OverallScore,
		ScoreFormulaVersion: review.ScoreFormulaVersion,
		Content:             review.Content,
		Tags:                canteenStringArray(review.Tags),
		Status:              string(review.Status),
		LikeCount:           review.LikeCount,
		DislikeCount:        review.DislikeCount,
		HiddenReason:        review.HiddenReason,
		HiddenBy:            review.HiddenBy,
		HiddenAt:            review.HiddenAt,
		CreatedAt:           review.CreatedAt,
		UpdatedAt:           review.UpdatedAt,
	}
	if review.ClientRequestID != "" {
		requestID := review.ClientRequestID
		entity.ClientRequestID = &requestID
	}
	return entity
}

func newCanteenReviewDomain(entity *CanteenReviewEventEntity) *canteen.ReviewEvent {
	if entity == nil {
		return nil
	}
	review := &canteen.ReviewEvent{
		ID:        entity.ID,
		StoreID:   entity.StoreID,
		UserID:    entity.UserID,
		VisitedAt: entity.VisitedAt,
		Dimensions: canteen.Dimensions{
			Taste:   entity.Taste,
			Value:   entity.Value,
			Queue:   entity.Queue,
			Hygiene: entity.Hygiene,
			Service: entity.Service,
		},
		OverallScore:        entity.OverallScore,
		ScoreFormulaVersion: entity.ScoreFormulaVersion,
		Content:             entity.Content,
		Tags:                []string(entity.Tags),
		Status:              canteen.ReviewStatus(entity.Status),
		LikeCount:           entity.LikeCount,
		DislikeCount:        entity.DislikeCount,
		HiddenReason:        entity.HiddenReason,
		HiddenBy:            entity.HiddenBy,
		HiddenAt:            entity.HiddenAt,
		CreatedAt:           entity.CreatedAt,
		UpdatedAt:           entity.UpdatedAt,
	}
	if entity.ClientRequestID != nil {
		review.ClientRequestID = *entity.ClientRequestID
	}
	return review
}

func newCanteenRevisionEntity(revision *canteen.Revision) CanteenReviewRevisionEntity {
	return CanteenReviewRevisionEntity{
		ID:            revision.ID,
		ReviewEventID: revision.ReviewEventID,
		CreatedBy:     revision.CreatedBy,
		VisitedAt:     revision.VisitedAt,
		Taste:         revision.Dimensions.Taste,
		Value:         revision.Dimensions.Value,
		Queue:         revision.Dimensions.Queue,
		Hygiene:       revision.Dimensions.Hygiene,
		Service:       revision.Dimensions.Service,
		OverallScore:  revision.OverallScore,
		Content:       revision.Content,
		Tags:          canteenStringArray(revision.Tags),
		CreatedAt:     revision.CreatedAt,
	}
}

func newCanteenUserSummaryDomain(entity *CanteenUserRatingSummaryEntity) *canteen.UserRatingSummary {
	if entity == nil {
		return nil
	}
	return &canteen.UserRatingSummary{
		StoreID:            entity.StoreID,
		UserID:             entity.UserID,
		WeightedScore:      entity.WeightedScore,
		WindowSize:         entity.WindowSize,
		LifetimeEventCount: entity.LifetimeEventCount,
		LastVisitedAt:      entity.LastVisitedAt,
		ComputedAt:         entity.ComputedAt,
	}
}

func newCanteenMediaDomain(entity *CanteenMediaEntity) *canteen.Media {
	if entity == nil {
		return nil
	}
	return &canteen.Media{
		ID:            entity.ID,
		OwnerUserID:   entity.OwnerUserID,
		Purpose:       canteen.MediaPurpose(entity.Purpose),
		StorageDriver: entity.StorageDriver,
		StorageKey:    entity.StorageKey,
		MIME:          entity.MIME,
		ByteSize:      entity.ByteSize,
		Width:         entity.Width,
		Height:        entity.Height,
		SHA256:        entity.SHA256,
		Status:        canteen.MediaStatus(entity.Status),
		ReviewedBy:    entity.ReviewedBy,
		ReviewedAt:    entity.ReviewedAt,
		RejectReason:  entity.RejectReason,
		ExpiresAt:     entity.ExpiresAt,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}
}

func newCanteenReportDomain(entity *CanteenReportEntity) *canteen.Report {
	if entity == nil {
		return nil
	}
	return &canteen.Report{
		ID:             entity.ID,
		ReporterUserID: entity.ReporterUserID,
		SubjectType:    canteen.ReportSubjectType(entity.SubjectType),
		SubjectID:      entity.SubjectID,
		ReasonCode:     canteen.ReportReason(entity.ReasonCode),
		Detail:         entity.Detail,
		Status:         canteen.ReportStatus(entity.Status),
		HandledBy:      entity.HandledBy,
		HandledAt:      entity.HandledAt,
		ResolutionNote: entity.ResolutionNote,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
	}
}
