package application

import (
	"context"
	"strconv"
	"time"

	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

// CanteenHallCommand is the editable hall payload.
type CanteenHallCommand struct {
	CampusKey       string
	Name            string
	ShortName       string
	Address         string
	Latitude        *float64
	Longitude       *float64
	ExternalPlaceID string
	Status          canteen.HallStatus
	DisplayOrder    int
}

// CanteenStoreCommand is the editable store payload.
type CanteenStoreCommand struct {
	HallID       int
	Name         string
	Floor        string
	Categories   []string
	Status       canteen.StoreStatus
	CoverMediaID *int
}

// CanteenAdminService maintains the canteen directory. It deliberately ignores
// the public canteen.enabled toggle so administrators can seed data before the
// feature is switched on.
type CanteenAdminService struct {
	halls  canteen.HallRepository
	stores canteen.StoreRepository
	now    func() time.Time
}

func NewCanteenAdminService(halls canteen.HallRepository, stores canteen.StoreRepository) *CanteenAdminService {
	return &CanteenAdminService{halls: halls, stores: stores, now: time.Now}
}

func (s *CanteenAdminService) CreateHall(ctx context.Context, actorID int, cmd CanteenHallCommand) (*canteen.Hall, error) {
	hall := &canteen.Hall{
		CampusKey:       cmd.CampusKey,
		Name:            cmd.Name,
		ShortName:       cmd.ShortName,
		Address:         cmd.Address,
		Latitude:        cmd.Latitude,
		Longitude:       cmd.Longitude,
		ExternalPlaceID: cmd.ExternalPlaceID,
		Status:          cmd.Status,
		DisplayOrder:    cmd.DisplayOrder,
		CreatedAt:       s.now(),
		UpdatedAt:       s.now(),
	}
	hall.ApplyDefaults()
	if err := hall.Validate(); err != nil {
		return nil, err
	}
	if err := s.halls.Create(ctx, hall); err != nil {
		return nil, err
	}
	s.record(ctx, actorID, audit.ActionCanteenHallCreate, audit.TargetTypeCanteenHall, hall.ID, audit.Details{
		"campus_key": hall.CampusKey,
		"name":       hall.Name,
	})
	return hall, nil
}

func (s *CanteenAdminService) UpdateHall(ctx context.Context, actorID int, hallID int, cmd CanteenHallCommand) (*canteen.Hall, error) {
	hall, err := s.halls.Get(ctx, hallID)
	if err != nil {
		return nil, err
	}
	if hall == nil {
		return nil, canteen.ErrHallNotFound
	}
	hall.CampusKey = cmd.CampusKey
	hall.Name = cmd.Name
	hall.ShortName = cmd.ShortName
	hall.Address = cmd.Address
	hall.Latitude = cmd.Latitude
	hall.Longitude = cmd.Longitude
	hall.ExternalPlaceID = cmd.ExternalPlaceID
	hall.Status = cmd.Status
	hall.DisplayOrder = cmd.DisplayOrder
	hall.ApplyDefaults()
	if err := hall.Validate(); err != nil {
		return nil, err
	}
	hall.UpdatedAt = s.now()
	if err := s.halls.Update(ctx, hall); err != nil {
		return nil, err
	}
	s.record(ctx, actorID, audit.ActionCanteenHallUpdate, audit.TargetTypeCanteenHall, hall.ID, audit.Details{
		"campus_key": hall.CampusKey,
		"name":       hall.Name,
		"status":     string(hall.Status),
	})
	return hall, nil
}

func (s *CanteenAdminService) CreateStore(ctx context.Context, actorID int, cmd CanteenStoreCommand) (*canteen.Store, error) {
	hall, err := s.halls.Get(ctx, cmd.HallID)
	if err != nil {
		return nil, err
	}
	if hall == nil {
		return nil, canteen.ErrHallNotFound
	}
	store := &canteen.Store{
		HallID:       cmd.HallID,
		Name:         cmd.Name,
		Floor:        cmd.Floor,
		Categories:   cmd.Categories,
		Status:       cmd.Status,
		CoverMediaID: cmd.CoverMediaID,
		CreatedAt:    s.now(),
		UpdatedAt:    s.now(),
	}
	store.ApplyDefaults()
	if err := store.Validate(); err != nil {
		return nil, err
	}
	if err := s.stores.Create(ctx, store); err != nil {
		return nil, err
	}
	s.record(ctx, actorID, audit.ActionCanteenStoreCreate, audit.TargetTypeCanteenStore, store.ID, audit.Details{
		"hall_id": store.HallID,
		"name":    store.Name,
		"status":  string(store.Status),
	})
	return store, nil
}

func (s *CanteenAdminService) UpdateStore(ctx context.Context, actorID int, storeID int, cmd CanteenStoreCommand) (*canteen.Store, error) {
	store, err := s.stores.Get(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, canteen.ErrStoreNotFound
	}
	if cmd.HallID != store.HallID {
		hall, err := s.halls.Get(ctx, cmd.HallID)
		if err != nil {
			return nil, err
		}
		if hall == nil {
			return nil, canteen.ErrHallNotFound
		}
	}
	store.HallID = cmd.HallID
	store.Name = cmd.Name
	store.Floor = cmd.Floor
	store.Categories = cmd.Categories
	store.Status = cmd.Status
	store.CoverMediaID = cmd.CoverMediaID
	store.ApplyDefaults()
	if err := store.Validate(); err != nil {
		return nil, err
	}
	store.UpdatedAt = s.now()
	if err := s.stores.Update(ctx, store); err != nil {
		return nil, err
	}
	s.record(ctx, actorID, audit.ActionCanteenStoreUpdate, audit.TargetTypeCanteenStore, store.ID, audit.Details{
		"hall_id": store.HallID,
		"name":    store.Name,
		"status":  string(store.Status),
	})
	return store, nil
}

func (s *CanteenAdminService) UpdateStoreStatus(ctx context.Context, actorID int, storeID int, status canteen.StoreStatus) (*canteen.Store, error) {
	if !status.Valid() {
		return nil, apperr.BadRequest("店铺状态无效")
	}
	store, err := s.stores.Get(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, canteen.ErrStoreNotFound
	}
	now := s.now()
	if err := store.TransitionTo(status, now); err != nil {
		return nil, err
	}
	if err := s.stores.UpdateStatus(ctx, store.ID, status, now); err != nil {
		return nil, err
	}
	s.record(ctx, actorID, audit.ActionCanteenStoreStatus, audit.TargetTypeCanteenStore, store.ID, audit.Details{
		"status": string(status),
	})
	return store, nil
}

func (s *CanteenAdminService) MergeStore(ctx context.Context, actorID int, storeID int, targetID int) (*canteen.Store, error) {
	if storeID == targetID {
		return nil, apperr.BadRequest("不能合并到自身")
	}
	store, err := s.stores.Get(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, canteen.ErrStoreNotFound
	}
	if store.Status == canteen.StoreStatusMerged {
		return nil, canteen.ErrStoreMerged
	}
	target, err := s.stores.Get(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, canteen.ErrStoreNotFound
	}
	now := s.now()
	if err := s.stores.Merge(ctx, store.ID, target.ID, now); err != nil {
		return nil, err
	}
	store.Status = canteen.StoreStatusMerged
	store.MergedIntoStoreID = &target.ID
	store.UpdatedAt = now
	s.record(ctx, actorID, audit.ActionCanteenStoreMerge, audit.TargetTypeCanteenStore, store.ID, audit.Details{
		"merged_into": target.ID,
	})
	return store, nil
}

func (s *CanteenAdminService) record(ctx context.Context, actorID int, action, targetType string, targetID int, details audit.Details) {
	if actorID <= 0 {
		return
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  s.now(),
		ActorUserID: actorID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    strconv.Itoa(targetID),
		Details:     details,
	})
}
