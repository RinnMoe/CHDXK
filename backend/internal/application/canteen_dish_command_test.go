package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

func newDishCommandFixture(t *testing.T) (*CanteenDishCommandService, *canteen.MockDishRepository, *canteen.MockDishReviewRepository, *canteen.MockDishQuery, *canteen.MockReviewQuery) {
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
	reviewQuery := &canteen.MockReviewQuery{}
	service := NewCanteenDishCommandService(
		stores,
		dishes,
		query,
		dishReviews,
		reviewQuery,
		StaticCanteenSettingsProvider{Runtime: CanteenRuntimeConfig{Enabled: true, ReviewsEnabled: true, MediaEnabled: true}},
		nil,
	)
	return service, dishes, dishReviews, query, reviewQuery
}

func TestCanteenDishCreatePendingAndDedupe(t *testing.T) {
	service, dishes, _, _, _ := newDishCommandFixture(t)
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
	service, _, _, _, _ := newDishCommandFixture(t)
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
	service, _, dishReviews, query, _ := newDishCommandFixture(t)
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
	service, _, _, query, _ := newDishCommandFixture(t)
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

func TestCanteenDishReviewRejectsForeignVisit(t *testing.T) {
	service, _, _, query, reviewQuery := newDishCommandFixture(t)
	query.Dish = &canteen.Dish{ID: 9, StoreID: 1, Name: "牛肉面", Status: canteen.DishStatusActive}
	reviewQuery.Reviews = []canteen.ReviewView{{
		ID: 7, StoreID: 1, UserID: 3, Status: canteen.ReviewStatusActive,
	}}
	user := &auth.User{ID: 2, Role: auth.RoleUser}
	if _, err := service.CreateDishReview(context.Background(), user, CreateCanteenDishReviewCommand{
		DishID: 9, ReviewEventID: intPtr(7), Taste: 5, Portion: 5, Value: 5,
	}); err == nil {
		t.Fatal("CreateDishReview should reject another user's visit")
	}
	reviewQuery.Reviews[0].UserID = 2
	if _, err := service.CreateDishReview(context.Background(), user, CreateCanteenDishReviewCommand{
		DishID: 9, ReviewEventID: intPtr(7), Taste: 5, Portion: 5, Value: 5,
	}); err != nil {
		t.Fatalf("CreateDishReview(own visit) = %v", err)
	}
}

func intPtr(value int) *int {
	return &value
}

func TestCanteenDishCreateRejectsMutedUser(t *testing.T) {
	service, _, _, _, _ := newDishCommandFixture(t)
	until := time.Now().Add(24 * time.Hour)
	muted := &auth.User{ID: 2, Role: auth.RoleUser, MutedUntil: &until}
	if _, _, err := service.CreateDish(context.Background(), muted, CreateCanteenDishCommand{StoreID: 1, Name: "牛肉面"}); !errors.Is(err, auth.ErrMuted) {
		t.Fatalf("CreateDish(muted) = %v, want ErrMuted", err)
	}
}
