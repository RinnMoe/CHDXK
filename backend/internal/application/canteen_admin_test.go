package application

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/domain/canteen"
)

func newCanteenAdminService(t *testing.T) (*CanteenAdminService, *canteen.MockHallRepository, *canteen.MockStoreRepository) {
	t.Helper()
	halls := canteen.NewMockHallRepository()
	stores := canteen.NewMockStoreRepository()
	if err := halls.Create(context.Background(), &canteen.Hall{
		CampusKey: "benbu",
		Name:      "一食堂",
		Status:    canteen.HallStatusActive,
	}); err != nil {
		t.Fatalf("seed hall: %v", err)
	}
	return NewCanteenAdminService(halls, stores), halls, stores
}

func TestCanteenAdminCreateStore(t *testing.T) {
	service, _, _ := newCanteenAdminService(t)
	store, err := service.CreateStore(context.Background(), 1, CanteenStoreCommand{
		HallID: 1,
		Name:   " 老王面馆 ",
		Status: canteen.StoreStatusActive,
	})
	if err != nil {
		t.Fatalf("CreateStore() = %v", err)
	}
	if store.ID == 0 {
		t.Fatal("store ID not assigned")
	}
	if store.NormalizedName != "老王面馆" || store.Name != "老王面馆" {
		t.Fatalf("store = %+v", store)
	}
}

func TestCanteenAdminCreateStoreMissingHall(t *testing.T) {
	service, _, _ := newCanteenAdminService(t)
	_, err := service.CreateStore(context.Background(), 1, CanteenStoreCommand{HallID: 99, Name: "店铺"})
	if !errors.Is(err, canteen.ErrHallNotFound) {
		t.Fatalf("CreateStore() err = %v, want ErrHallNotFound", err)
	}
}

func TestCanteenAdminStoreStatusTransition(t *testing.T) {
	service, _, _ := newCanteenAdminService(t)
	store, err := service.CreateStore(context.Background(), 1, CanteenStoreCommand{HallID: 1, Name: "老王面馆"})
	if err != nil {
		t.Fatalf("CreateStore() = %v", err)
	}
	if store.Status != canteen.StoreStatusPending {
		t.Fatalf("default status = %q, want pending", store.Status)
	}
	activated, err := service.UpdateStoreStatus(context.Background(), 1, store.ID, canteen.StoreStatusActive)
	if err != nil {
		t.Fatalf("UpdateStoreStatus(active) = %v", err)
	}
	if activated.Status != canteen.StoreStatusActive {
		t.Fatalf("status = %q", activated.Status)
	}
	if _, err := service.UpdateStoreStatus(context.Background(), 1, store.ID, canteen.StoreStatusRejected); err == nil {
		t.Fatal("active -> rejected must be rejected")
	}
}

func TestCanteenAdminMergeStore(t *testing.T) {
	service, _, _ := newCanteenAdminService(t)
	source, err := service.CreateStore(context.Background(), 1, CanteenStoreCommand{HallID: 1, Name: "老王面馆"})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	target, err := service.CreateStore(context.Background(), 1, CanteenStoreCommand{HallID: 1, Name: "老王面馆二店"})
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	merged, err := service.MergeStore(context.Background(), 1, source.ID, target.ID)
	if err != nil {
		t.Fatalf("MergeStore() = %v", err)
	}
	if merged.Status != canteen.StoreStatusMerged || merged.MergedIntoStoreID == nil || *merged.MergedIntoStoreID != target.ID {
		t.Fatalf("merged = %+v", merged)
	}
	if _, err := service.MergeStore(context.Background(), 1, source.ID, target.ID); !errors.Is(err, canteen.ErrStoreMerged) {
		t.Fatalf("second merge err = %v, want ErrStoreMerged", err)
	}
	if _, err := service.MergeStore(context.Background(), 1, target.ID, target.ID); err == nil {
		t.Fatal("self merge must be rejected")
	}
}
