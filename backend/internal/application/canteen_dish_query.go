package application

import (
	"context"
	"sort"

	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

// CanteenDishQueryService exposes dish lists, suggestions and contributions.
type CanteenDishQueryService struct {
	dishes   canteen.DishQuery
	settings CanteenSettingsProvider
}

func NewCanteenDishQueryService(dishes canteen.DishQuery, settings CanteenSettingsProvider) *CanteenDishQueryService {
	return &CanteenDishQueryService{dishes: dishes, settings: settings}
}

func (s *CanteenDishQueryService) requireEnabled(ctx context.Context) error {
	if s.settings == nil {
		return nil
	}
	config, err := s.settings.CanteenRuntimeConfig(ctx)
	if err != nil {
		return err
	}
	if !config.Enabled {
		return apperr.NotFound("食堂评价暂未开放")
	}
	return nil
}

func (s *CanteenDishQueryService) ListStoreDishes(ctx context.Context, storeID int) ([]CanteenDishDTO, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	items, err := s.dishes.ListDishes(ctx, canteen.DishListFilter{
		StoreID: storeID,
		Status:  canteen.DishStatusActive,
		Limit:   100,
	})
	if err != nil {
		return nil, err
	}
	return newCanteenDishDTOs(items), nil
}

func (s *CanteenDishQueryService) ListDishesForAdmin(ctx context.Context, filter canteen.DishListFilter) ([]CanteenDishDTO, error) {
	items, err := s.dishes.ListDishes(ctx, filter)
	if err != nil {
		return nil, err
	}
	return newCanteenDishDTOs(items), nil
}

func (s *CanteenDishQueryService) Suggest(ctx context.Context, storeID int, query string) ([]CanteenDishDTO, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	items, err := s.dishes.ListDishes(ctx, canteen.DishListFilter{
		StoreID: storeID,
		Status:  canteen.DishStatusActive,
		Q:       query,
		Limit:   canteen.MaxDishSuggestions,
	})
	if err != nil {
		return nil, err
	}
	// Exact/alias search is authoritative. Only when it is thin do we add
	// edit-distance candidates, and those remain suggestions, never merges.
	if len(items) < canteen.MaxDishSuggestions {
		all, err := s.dishes.ListDishes(ctx, canteen.DishListFilter{
			StoreID: storeID,
			Status:  canteen.DishStatusActive,
			Limit:   100,
		})
		if err != nil {
			return nil, err
		}
		seen := make(map[int]struct{}, len(items))
		for _, item := range items {
			seen[item.ID] = struct{}{}
		}
		type candidate struct {
			dish     canteen.DishView
			distance int
		}
		var fuzzy []candidate
		for _, item := range all {
			if _, ok := seen[item.ID]; ok {
				continue
			}
			distance := canteen.NameDistance(query, item.Name)
			if distance <= 2 {
				fuzzy = append(fuzzy, candidate{dish: item, distance: distance})
			}
		}
		sort.Slice(fuzzy, func(i, j int) bool {
			if fuzzy[i].distance != fuzzy[j].distance {
				return fuzzy[i].distance < fuzzy[j].distance
			}
			if fuzzy[i].dish.RatingScore != fuzzy[j].dish.RatingScore {
				return fuzzy[i].dish.RatingScore > fuzzy[j].dish.RatingScore
			}
			return fuzzy[i].dish.ID < fuzzy[j].dish.ID
		})
		for _, item := range fuzzy {
			if len(items) >= canteen.MaxDishSuggestions {
				break
			}
			items = append(items, item.dish)
		}
	}
	return newCanteenDishDTOs(items), nil
}

// GetDishDetail returns one dish with its approved photo gallery and recent
// independent dish ratings.
func (s *CanteenDishQueryService) GetDishDetail(ctx context.Context, storeID, dishID int) (*CanteenDishDetailDTO, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	dish, err := s.dishes.GetDish(ctx, dishID)
	if err != nil {
		return nil, err
	}
	if dish == nil || dish.Status != canteen.DishStatusActive || dish.StoreID != storeID {
		return nil, canteen.ErrDishNotFound
	}
	views, err := s.dishes.ListDishes(ctx, canteen.DishListFilter{
		StoreID: storeID,
		Status:  canteen.DishStatusActive,
		Limit:   100,
	})
	if err != nil {
		return nil, err
	}
	var dishView *canteen.DishView
	for i := range views {
		if views[i].ID == dishID {
			dishView = &views[i]
			break
		}
	}
	if dishView == nil {
		return nil, canteen.ErrDishNotFound
	}
	photos, err := s.dishes.ListDishPhotos(ctx, dishID)
	if err != nil {
		return nil, err
	}
	reviews, total, err := s.dishes.ListDishReviews(ctx, dishID, 1, 20)
	if err != nil {
		return nil, err
	}
	media := make([]CanteenMediaDTO, 0, len(photos))
	for _, item := range photos {
		media = append(media, CanteenMediaDTO{ID: item.ID, URL: item.URL, Width: item.Width, Height: item.Height})
	}
	reviewDTOs := make([]CanteenDishReviewDTO, 0, len(reviews))
	for _, item := range reviews {
		reviewDTOs = append(reviewDTOs, newCanteenDishReviewDTO(item))
	}
	return &CanteenDishDetailDTO{
		Dish:        newCanteenDishDTO(*dishView),
		Photos:      media,
		Reviews:     reviewDTOs,
		ReviewTotal: total,
	}, nil
}

func (s *CanteenDishQueryService) ListMyDishReviews(ctx context.Context, userID int, page, pageSize int) ([]CanteenDishReviewDTO, int64, int, int, error) {
	page, pageSize = normalizeCanteenPagination(page, pageSize)
	items, total, err := s.dishes.ListDishReviewsByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	result := make([]CanteenDishReviewDTO, 0, len(items))
	for _, item := range items {
		result = append(result, newCanteenDishReviewDTO(item))
	}
	return result, total, page, pageSize, nil
}

func newCanteenDishDTOs(items []canteen.DishView) []CanteenDishDTO {
	result := make([]CanteenDishDTO, 0, len(items))
	for _, item := range items {
		result = append(result, newCanteenDishDTO(item))
	}
	return result
}
