package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

func TestCanteenCatalogDisabled(t *testing.T) {
	catalog := &canteen.MockCatalog{}
	service := NewCanteenCatalogQueryService(catalog, catalog, StaticCanteenSettingsProvider{
		Runtime: CanteenRuntimeConfig{Enabled: false},
	})
	_, err := service.Home(context.Background())
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) || appErr.StatusCode != 404 {
		t.Fatalf("Home() err = %v, want 404", err)
	}
}

func TestCanteenCatalogHomeMapping(t *testing.T) {
	now := time.Now()
	catalog := &canteen.MockCatalog{Home: &canteen.HomeView{
		Halls: []canteen.HallView{{ID: 3, CampusKey: "benbu", Name: "一食堂", Status: canteen.HallStatusActive, StoreCount: 2}},
		TopStores: []canteen.StoreListItem{{
			ID: 5, HallID: 3, HallName: "一食堂", CampusKey: "benbu", Name: "老王面馆",
			Categories: []string{"面食"}, Status: canteen.StoreStatusActive,
			RatingCount: 4, RatingAvg: 4.25, RatingScore: 4.1, ReviewEventCount: 9, LastReviewAt: &now,
		}},
		NewStores: []canteen.StoreListItem{{ID: 6, Name: "新店", Status: canteen.StoreStatusActive}},
		LatestReviews: []canteen.ReviewView{{
			ID: 7, StoreID: 5, StoreName: "老王面馆", UserID: 2,
			Dimensions:   canteen.Dimensions{Taste: 5, Value: 4, Queue: 3, Hygiene: 4, Service: 5},
			OverallScore: 4, Content: "不错", Status: canteen.ReviewStatusActive, CreatedAt: now, UpdatedAt: now,
		}},
		Stats: canteen.HomeStats{StoreCount: 2, ReviewCount: 9, DishCount: 3},
	}}
	service := NewCanteenCatalogQueryService(catalog, catalog, StaticCanteenSettingsProvider{
		Runtime: CanteenRuntimeConfig{Enabled: true},
	})

	home, err := service.Home(context.Background())
	if err != nil {
		t.Fatalf("Home() = %v", err)
	}
	if len(home.Halls) != 1 || home.Halls[0].StoreCount != 2 {
		t.Fatalf("halls = %+v", home.Halls)
	}
	if len(home.TopStores) != 1 || home.TopStores[0].Rating.Score != 4.1 {
		t.Fatalf("top stores = %+v", home.TopStores)
	}
	if home.TopStores[0].ReviewCount != 9 || home.TopStores[0].CoverURL != "" {
		t.Fatalf("top store = %+v", home.TopStores[0])
	}
	if len(home.LatestReviews) != 1 || home.LatestReviews[0].Taste != 5 {
		t.Fatalf("reviews = %+v", home.LatestReviews)
	}
	if home.Stats != (CanteenHomeStatsDTO{StoreCount: 2, ReviewCount: 9, DishCount: 3}) {
		t.Fatalf("stats = %+v", home.Stats)
	}
}

func TestCanteenCatalogPaginationNormalized(t *testing.T) {
	catalog := &canteen.MockCatalog{
		Stores:     []canteen.StoreListItem{{ID: 1, Name: "店铺", Status: canteen.StoreStatusActive}},
		StoreTotal: 1,
	}
	service := NewCanteenCatalogQueryService(catalog, catalog, StaticCanteenSettingsProvider{
		Runtime: CanteenRuntimeConfig{Enabled: true},
	})
	page, err := service.ListStores(context.Background(), canteen.StoreListFilter{Page: 0, PageSize: 500})
	if err != nil {
		t.Fatalf("ListStores() = %v", err)
	}
	if page.Page != 1 || page.PageSize != 100 {
		t.Fatalf("page = %+v", page)
	}
	if len(page.Items) != 1 || page.Items[0].Status != string(canteen.StoreStatusActive) {
		t.Fatalf("items = %+v", page.Items)
	}
}

func TestCanteenCatalogStoreNotFound(t *testing.T) {
	catalog := &canteen.MockCatalog{}
	service := NewCanteenCatalogQueryService(catalog, catalog, StaticCanteenSettingsProvider{
		Runtime: CanteenRuntimeConfig{Enabled: true},
	})
	_, err := service.GetStore(context.Background(), 9, 2)
	if !errors.Is(err, canteen.ErrStoreNotFound) {
		t.Fatalf("GetStore() err = %v, want ErrStoreNotFound", err)
	}
}

func TestCanteenCatalogPendingStoreIsHidden(t *testing.T) {
	catalog := &canteen.MockCatalog{Detail: &canteen.StoreDetail{
		StoreListItem: canteen.StoreListItem{ID: 5, Status: canteen.StoreStatusPending},
	}}
	service := NewCanteenCatalogQueryService(catalog, catalog, StaticCanteenSettingsProvider{
		Runtime: CanteenRuntimeConfig{Enabled: true},
	})
	_, err := service.GetStore(context.Background(), 5, 2)
	if !errors.Is(err, canteen.ErrStoreNotFound) {
		t.Fatalf("GetStore(pending) err = %v, want ErrStoreNotFound", err)
	}
}
