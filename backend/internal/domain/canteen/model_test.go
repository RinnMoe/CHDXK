package canteen

import (
	"testing"
	"time"
)

func TestCanTransitionStore(t *testing.T) {
	cases := []struct {
		from StoreStatus
		to   StoreStatus
		want bool
	}{
		{StoreStatusPending, StoreStatusActive, true},
		{StoreStatusPending, StoreStatusRejected, true},
		{StoreStatusPending, StoreStatusMerged, true},
		{StoreStatusPending, StoreStatusOffline, false},
		{StoreStatusRejected, StoreStatusActive, true},
		{StoreStatusActive, StoreStatusOffline, true},
		{StoreStatusOffline, StoreStatusActive, true},
		{StoreStatusActive, StoreStatusRejected, false},
		{StoreStatusMerged, StoreStatusActive, false},
	}
	for _, tc := range cases {
		if got := CanTransitionStore(tc.from, tc.to); got != tc.want {
			t.Fatalf("CanTransitionStore(%s, %s) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestCanTransitionDish(t *testing.T) {
	cases := []struct {
		from DishStatus
		to   DishStatus
		want bool
	}{
		{DishStatusPending, DishStatusActive, true},
		{DishStatusPending, DishStatusRejected, true},
		{DishStatusActive, DishStatusHidden, true},
		{DishStatusHidden, DishStatusActive, true},
		{DishStatusActive, DishStatusPending, false},
		{DishStatusMerged, DishStatusHidden, false},
	}
	for _, tc := range cases {
		if got := CanTransitionDish(tc.from, tc.to); got != tc.want {
			t.Fatalf("CanTransitionDish(%s, %s) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestStoreApplyDefaultsAndValidate(t *testing.T) {
	store := Store{HallID: 1, Name: " 老王面馆 ", Categories: []string{" 面食 ", "面食", ""}}
	store.ApplyDefaults()
	if store.NormalizedName != "老王面馆" {
		t.Fatalf("NormalizedName = %q", store.NormalizedName)
	}
	if store.Status != StoreStatusPending {
		t.Fatalf("Status = %q, want pending", store.Status)
	}
	if len(store.Categories) != 1 || store.Categories[0] != "面食" {
		t.Fatalf("Categories = %v", store.Categories)
	}
	if err := store.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}

	invalid := Store{Name: "面馆"}
	if err := invalid.Validate(); err == nil {
		t.Fatal("Validate() should reject missing hall")
	}
}

func TestStoreVisibility(t *testing.T) {
	active := Store{Status: StoreStatusActive}
	if !active.CanAcceptReview() || !active.PubliclyVisible() {
		t.Fatal("active store must accept reviews and be visible")
	}
	offline := Store{Status: StoreStatusOffline}
	if offline.CanAcceptReview() {
		t.Fatal("offline store must not accept reviews")
	}
	if !offline.PubliclyVisible() {
		t.Fatal("offline store keeps its history visible")
	}
	pending := Store{Status: StoreStatusPending}
	if pending.PubliclyVisible() {
		t.Fatal("pending store must not be public")
	}
}

func TestStoreTransitionTo(t *testing.T) {
	store := Store{Status: StoreStatusPending}
	now := time.Now()
	if err := store.TransitionTo(StoreStatusActive, now); err != nil {
		t.Fatalf("TransitionTo(active) = %v", err)
	}
	if store.Status != StoreStatusActive || !store.UpdatedAt.Equal(now) {
		t.Fatalf("store = %+v", store)
	}
	if err := store.TransitionTo(StoreStatusRejected, now); err == nil {
		t.Fatal("active -> rejected should be rejected")
	}
}

func TestDishApplyDefaultsAndValidate(t *testing.T) {
	dish := Dish{StoreID: 1, Name: " 牛肉面 "}
	dish.ApplyDefaults()
	if dish.NormalizedName != "牛肉面" || dish.Status != DishStatusPending {
		t.Fatalf("dish = %+v", dish)
	}
	if err := dish.Validate(); err != nil {
		t.Fatalf("Validate() = %v", err)
	}
	if dish.PubliclyVisible() {
		t.Fatal("pending dish must not be public")
	}
	dish.Status = DishStatusActive
	if !dish.PubliclyVisible() {
		t.Fatal("active dish must be public")
	}
}
