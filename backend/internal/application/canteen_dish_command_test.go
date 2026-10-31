package application

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

func newDishCommandFixture(t *testing.T) (*CanteenDishCommandService, *canteen.MockDishRepository, *canteen.MockDishReviewRepository, *canteen.MockDishQuery) {
	t.Helper()
	stores := canteen.NewMockStoreRepository()
	store := &canteen.Store{HallID: 1, Name: "老王面馆", Status: canteen.StoreStatusActive}
	store.ApplyDefaults()
	if err := stores.Create(context.Background(), store); err != nil {
		t.Fatalf("seed store: %v", err)
	}
	dishes := canteen.NewMockDishRepository()
	dishReviews := canteen.NewMockDishReviewRepository()
	query := &canteen.MockDishQuery{}
	service := NewCanteenDishCommandService(
		stores,
		dishes,
		query,
		dishReviews,
		StaticCanteenSettingsProvider{Runtime: CanteenRuntimeConfig{Enabled: true, ReviewsEnabled: true, MediaEnabled: true}},
		nil,
	)
	return service, dishes, dishReviews, query
}

func TestCanteenDishCreatePendingAndDedupe(t *testing.T) {
	service, dishes, _, _ := newDishCommandFixture(t)
	user := &auth.User{ID: 2, Role: auth.RoleUser}
	dish, existing, err := service.CreateDish(context.Background(), user, CreateCanteenDishCommand{StoreID: 1, Name: " 牛肉面 "})
	if err != nil {
		t.Fatalf("CreateDish() = %v", err)
	}
	if existing || dish.Status != canteen.DishStatusPending || dish.NormalizedName != "牛肉面" {
		t.Fatalf("dish = %+v existing=%v", dish, existing)
	}
	duplicate, existing, err := service.CreateDish(context.Background(), user, CreateCanteenDishCommand{StoreID: 1, Name: "牛肉面"})
	if err != nil {
		t.Fatalf("CreateDish(duplicate) = %v", err)
	}
	if !existing || duplicate.ID != dish.ID {
		t.Fatalf("duplicate = %+v existing=%v", duplicate, existing)
	}
	if len(dishes.Dishes) != 1 {
		t.Fatalf("dish rows = %d", len(dishes.Dishes))
	}
}

func TestCanteenDishAdminCreateIsActive(t *testing.T) {
	service, _, _, _ := newDishCommandFixture(t)
	admin := &auth.User{ID: 1, Role: auth.RoleAdmin}
	dish, existing, err := service.CreateDish(context.Background(), admin, CreateCanteenDishCommand{StoreID: 1, Name: "牛肉面"})
	if err != nil {
		t.Fatalf("CreateDish(admin) = %v", err)
	}
	if existing || dish.Status != canteen.DishStatusActive {
		t.Fatalf("dish = %+v existing=%v", dish, existing)
	}
}

func TestCanteenDishReview(t *testing.T) {
	service, _, dishReviews, query := newDishCommandFixture(t)
	query.Dish = &canteen.Dish{ID: 9, StoreID: 1, Name: "牛肉面", Status: canteen.DishStatusActive}
	user := &auth.User{ID: 2, Role: auth.RoleUser}
	review, err := service.CreateDishReview(context.Background(), user, CreateCanteenDishReviewCommand{
		DishID: 9, Taste: 5, Portion: 4, Value: 3, Content: "好吃",
	})
	if err != nil {
		t.Fatalf("CreateDishReview() = %v", err)
	}
	if review.ID == 0 || review.OverallScore != 4.25 || len(dishReviews.Reviews) != 1 {
		t.Fatalf("review = %+v", review)
	}
	query.Dish = &canteen.Dish{ID: 9, StoreID: 1, Name: "牛肉面", Status: canteen.DishStatusPending}
	if _, err := service.CreateDishReview(context.Background(), user, CreateCanteenDishReviewCommand{
		DishID: 9, Taste: 5, Portion: 5, Value: 5,
	}); !errors.Is(err, canteen.ErrDishNotFound) {
		t.Fatalf("CreateDishReview(pending) = %v", err)
	}
}

func TestCanteenDishAuthority(t *testing.T) {
	service, _, _, query := newDishCommandFixture(t)
	query.Dish = &canteen.Dish{ID: 9, StoreID: 1, Name: "牛肉面", Status: canteen.DishStatusActive}
	user := &auth.User{ID: 2, Role: auth.RoleUser}
	if _, err := service.UpdateDishStatus(context.Background(), user, 9, canteen.DishStatusActive); !errors.Is(err, canteen.ErrCannotModerate) {
		t.Fatalf("UpdateDishStatus(user) = %v", err)
	}
	if err := service.MergeDish(context.Background(), user, 9, 10); !errors.Is(err, canteen.ErrCannotModerate) {
		t.Fatalf("MergeDish(user) = %v", err)
	}
	admin := &auth.User{ID: 1, Role: auth.RoleAdmin}
	if err := service.MergeDish(context.Background(), admin, 9, 9); err == nil {
		t.Fatal("self merge should be rejected")
	}
}
