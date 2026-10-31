package application

import (
	"context"
	"strings"

	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

// CanteenSearchQueryService provides one grouped query for canteens and dishes.
type CanteenSearchQueryService struct {
	stores   canteen.StoreQuery
	dishes   canteen.DishQuery
	settings CanteenSettingsProvider
}

func NewCanteenSearchQueryService(stores canteen.StoreQuery, dishes canteen.DishQuery, settings CanteenSettingsProvider) *CanteenSearchQueryService {
	return &CanteenSearchQueryService{stores: stores, dishes: dishes, settings: settings}
}

func (s *CanteenSearchQueryService) Search(ctx context.Context, query string) (*CanteenSearchDTO, error) {
	if s.settings != nil {
		config, err := s.settings.CanteenRuntimeConfig(ctx)
		if err != nil {
			return nil, err
		}
		if !config.Enabled {
			return nil, apperr.NotFound("食堂评价暂未开放")
		}
	}
	query = strings.TrimSpace(query)
	result := &CanteenSearchDTO{
		Stores: []CanteenStoreListItemDTO{},
		Dishes: []CanteenDishDTO{},
	}
	if query == "" {
		return result, nil
	}
	stores, _, err := s.stores.ListStores(ctx, canteen.StoreListFilter{
		Q:        query,
		Status:   canteen.StoreStatusActive,
		Sort:     "rating_score",
		Page:     1,
		PageSize: 6,
	})
	if err != nil {
		return nil, err
	}
	for _, store := range stores {
		result.Stores = append(result.Stores, newCanteenStoreListItemDTO(store))
	}
	dishes, err := s.dishes.ListDishes(ctx, canteen.DishListFilter{
		Status: canteen.DishStatusActive,
		Q:      query,
		Limit:  8,
	})
	if err != nil {
		return nil, err
	}
	for _, dish := range dishes {
		result.Dishes = append(result.Dishes, newCanteenDishDTO(dish))
	}
	return result, nil
}
