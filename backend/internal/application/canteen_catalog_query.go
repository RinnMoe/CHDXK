package application

import (
	"context"

	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

const (
	defaultCanteenPage     = 1
	defaultCanteenPageSize = 20
	maxCanteenPageSize     = 100
)

// CanteenCatalogQueryService exposes the public canteen directory.
type CanteenCatalogQueryService struct {
	halls    canteen.HallQuery
	stores   canteen.StoreQuery
	settings CanteenSettingsProvider
}

func NewCanteenCatalogQueryService(
	halls canteen.HallQuery,
	stores canteen.StoreQuery,
	settings CanteenSettingsProvider,
) *CanteenCatalogQueryService {
	return &CanteenCatalogQueryService{halls: halls, stores: stores, settings: settings}
}

func (s *CanteenCatalogQueryService) runtimeEnabled(ctx context.Context) (bool, error) {
	if s.settings == nil {
		return true, nil
	}
	config, err := s.settings.CanteenRuntimeConfig(ctx)
	if err != nil {
		return false, err
	}
	return config.Enabled, nil
}

func (s *CanteenCatalogQueryService) requireEnabled(ctx context.Context) error {
	enabled, err := s.runtimeEnabled(ctx)
	if err != nil {
		return err
	}
	if !enabled {
		return apperr.NotFound("食堂评价暂未开放")
	}
	return nil
}

func (s *CanteenCatalogQueryService) ListHalls(ctx context.Context) ([]CanteenHallDTO, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	halls, err := s.halls.ListHalls(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]CanteenHallDTO, 0, len(halls))
	for _, hall := range halls {
		items = append(items, newCanteenHallDTO(hall))
	}
	return items, nil
}

func (s *CanteenCatalogQueryService) ListStores(ctx context.Context, filter canteen.StoreListFilter) (*CanteenStorePage, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	filter.Page, filter.PageSize = normalizeCanteenPagination(filter.Page, filter.PageSize)
	filter.Status = canteen.StoreStatusActive
	return s.listStores(ctx, filter)
}

func (s *CanteenCatalogQueryService) ListStoresForAdmin(ctx context.Context, filter canteen.StoreListFilter) (*CanteenStorePage, error) {
	filter.Page, filter.PageSize = normalizeCanteenPagination(filter.Page, filter.PageSize)
	return s.listStores(ctx, filter)
}

func (s *CanteenCatalogQueryService) listStores(ctx context.Context, filter canteen.StoreListFilter) (*CanteenStorePage, error) {
	items, total, err := s.stores.ListStores(ctx, filter)
	if err != nil {
		return nil, err
	}
	page := make([]CanteenStoreListItemDTO, 0, len(items))
	for _, item := range items {
		page = append(page, newCanteenStoreListItemDTO(item))
	}
	return &CanteenStorePage{
		Items:    page,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *CanteenCatalogQueryService) GetStore(ctx context.Context, storeID int, viewerID int) (*CanteenStoreDetailDTO, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	detail, err := s.stores.GetStoreDetail(ctx, storeID, viewerID)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, canteen.ErrStoreNotFound
	}
	if !detail.StoreListItem.PubliclyVisible() {
		return nil, canteen.ErrStoreNotFound
	}
	return newCanteenStoreDetailDTO(detail), nil
}

func (s *CanteenCatalogQueryService) ListRankings(ctx context.Context, filter canteen.RankingFilter) (*CanteenStorePage, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	filter.Page, filter.PageSize = normalizeCanteenPagination(filter.Page, filter.PageSize)
	items, total, err := s.stores.ListRankings(ctx, filter)
	if err != nil {
		return nil, err
	}
	page := make([]CanteenStoreListItemDTO, 0, len(items))
	for _, item := range items {
		page = append(page, newCanteenStoreListItemDTO(item))
	}
	return &CanteenStorePage{
		Items:    page,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *CanteenCatalogQueryService) Home(ctx context.Context) (*CanteenHomeDTO, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	home, err := s.stores.GetHome(ctx)
	if err != nil {
		return nil, err
	}
	result := &CanteenHomeDTO{
		Halls:         make([]CanteenHallDTO, 0, len(home.Halls)),
		TopStores:     make([]CanteenStoreListItemDTO, 0, len(home.TopStores)),
		NewStores:     make([]CanteenStoreListItemDTO, 0, len(home.NewStores)),
		LatestReviews: make([]CanteenReviewDTO, 0, len(home.LatestReviews)),
		HotDishes:     make([]CanteenDishDTO, 0, len(home.HotDishes)),
		Feeds:         make([]CanteenHomeFeedDTO, 0, len(home.Feeds)),
		Stats: CanteenHomeStatsDTO{
			StoreCount:  home.Stats.StoreCount,
			ReviewCount: home.Stats.ReviewCount,
			DishCount:   home.Stats.DishCount,
		},
	}
	for _, hall := range home.Halls {
		result.Halls = append(result.Halls, newCanteenHallDTO(hall))
	}
	for _, item := range home.TopStores {
		result.TopStores = append(result.TopStores, newCanteenStoreListItemDTO(item))
	}
	for _, item := range home.NewStores {
		result.NewStores = append(result.NewStores, newCanteenStoreListItemDTO(item))
	}
	for _, review := range home.LatestReviews {
		result.LatestReviews = append(result.LatestReviews, newCanteenReviewDTO(review))
	}
	for _, dish := range home.HotDishes {
		result.HotDishes = append(result.HotDishes, newCanteenDishDTO(dish))
	}
	if home.Hero != nil {
		result.Hero = &CanteenHomeHeroDTO{
			Type:        home.Hero.Type,
			StoreID:     home.Hero.StoreID,
			StoreName:   home.Hero.StoreName,
			HallName:    home.Hero.HallName,
			DishID:      home.Hero.DishID,
			DishName:    home.Hero.DishName,
			ImageURL:    home.Hero.ImageURL,
			Reason:      home.Hero.Reason,
			RatingScore: home.Hero.RatingScore,
			RatingCount: home.Hero.RatingCount,
		}
	}
	for _, feed := range home.Feeds {
		result.Feeds = append(result.Feeds, CanteenHomeFeedDTO{
			Type:        feed.Type,
			StoreID:     feed.StoreID,
			StoreName:   feed.StoreName,
			HallName:    feed.HallName,
			DishID:      feed.DishID,
			DishName:    feed.DishName,
			Reason:      feed.Reason,
			ImageURL:    feed.ImageURL,
			RatingScore: feed.RatingScore,
			RatingCount: feed.RatingCount,
			Tags:        append([]string{}, feed.Tags...),
		})
	}
	return result, nil
}

func newCanteenStoreDetailDTO(detail *canteen.StoreDetail) *CanteenStoreDetailDTO {
	result := &CanteenStoreDetailDTO{
		CanteenStoreListItemDTO: newCanteenStoreListItemDTO(detail.StoreListItem),
		Dimensions: CanteenDimensionsDTO{
			Taste:   detail.Dimensions.Taste,
			Value:   detail.Dimensions.Value,
			Queue:   detail.Dimensions.Queue,
			Hygiene: detail.Dimensions.Hygiene,
			Service: detail.Dimensions.Service,
		},
		Dishes: make([]CanteenDishDTO, 0, len(detail.Dishes)),
	}
	for _, dish := range detail.Dishes {
		result.Dishes = append(result.Dishes, newCanteenDishDTO(dish))
	}
	if detail.MySummary != nil {
		result.MySummary = &CanteenMySummaryDTO{
			WeightedScore:      detail.MySummary.WeightedScore,
			WindowSize:         detail.MySummary.WindowSize,
			LifetimeEventCount: detail.MySummary.LifetimeEventCount,
			LastVisitedAt:      detail.MySummary.LastVisitedAt,
		}
	}
	if detail.MyLatestReview != nil {
		review := newCanteenReviewDTO(*detail.MyLatestReview)
		result.MyLatestReview = &review
	}
	return result
}

func normalizeCanteenPagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultCanteenPage
	}
	if pageSize <= 0 {
		pageSize = defaultCanteenPageSize
	}
	if pageSize > maxCanteenPageSize {
		pageSize = maxCanteenPageSize
	}
	return page, pageSize
}
