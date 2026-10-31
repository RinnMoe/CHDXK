package canteen

import (
	"strings"
	"time"
	"unicode/utf8"

	"jcourse/pkg/apperr"
)

const (
	MaxHallNameRunes   = 60
	MaxAddressRunes    = 200
	MaxStoreNameRunes  = 60
	MaxFloorRunes      = 40
	MaxStoreCategories = 8
	MaxCategoryRunes   = 20
	MaxDishNameRunes   = 60
)

type HallStatus string

const (
	HallStatusActive  HallStatus = "active"
	HallStatusOffline HallStatus = "offline"
	HallStatusHidden  HallStatus = "hidden"
)

func (s HallStatus) Valid() bool {
	switch s {
	case HallStatusActive, HallStatusOffline, HallStatusHidden:
		return true
	default:
		return false
	}
}

// Hall is a canteen building on a campus. Campus is intentionally a free-form
// key managed by administrators instead of a foreign key to an external map.
type Hall struct {
	ID              int
	CampusKey       string
	Name            string
	ShortName       string
	Address         string
	Latitude        *float64
	Longitude       *float64
	ExternalPlaceID string
	Status          HallStatus
	DisplayOrder    int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (h *Hall) ApplyDefaults() {
	h.CampusKey = strings.TrimSpace(h.CampusKey)
	h.Name = strings.TrimSpace(h.Name)
	h.ShortName = strings.TrimSpace(h.ShortName)
	h.Address = strings.TrimSpace(h.Address)
	h.ExternalPlaceID = strings.TrimSpace(h.ExternalPlaceID)
	if h.Status == "" {
		h.Status = HallStatusActive
	}
}

func (h *Hall) Validate() error {
	if h.CampusKey == "" || h.Name == "" {
		return ErrHallInvalid
	}
	if utf8.RuneCountInString(h.Name) > MaxHallNameRunes {
		return apperr.BadRequest("食堂名称过长")
	}
	if utf8.RuneCountInString(h.Address) > MaxAddressRunes {
		return apperr.BadRequest("食堂地址过长")
	}
	if !h.Status.Valid() {
		return apperr.BadRequest("食堂状态无效")
	}
	if h.Latitude != nil && (*h.Latitude < -90 || *h.Latitude > 90) {
		return apperr.BadRequest("纬度超出范围")
	}
	if h.Longitude != nil && (*h.Longitude < -180 || *h.Longitude > 180) {
		return apperr.BadRequest("经度超出范围")
	}
	return nil
}

type StoreStatus string

const (
	StoreStatusPending  StoreStatus = "pending"
	StoreStatusActive   StoreStatus = "active"
	StoreStatusOffline  StoreStatus = "offline"
	StoreStatusRejected StoreStatus = "rejected"
	StoreStatusMerged   StoreStatus = "merged"
)

func (s StoreStatus) Valid() bool {
	switch s {
	case StoreStatusPending, StoreStatusActive, StoreStatusOffline, StoreStatusRejected, StoreStatusMerged:
		return true
	default:
		return false
	}
}

var storeTransitions = map[StoreStatus][]StoreStatus{
	StoreStatusPending:  {StoreStatusActive, StoreStatusRejected, StoreStatusMerged},
	StoreStatusRejected: {StoreStatusActive, StoreStatusMerged},
	StoreStatusActive:   {StoreStatusOffline, StoreStatusMerged},
	StoreStatusOffline:  {StoreStatusActive, StoreStatusMerged},
	StoreStatusMerged:   {},
}

// CanTransitionStore reports whether an administrator may move a store between
// the two states. Merged stores are terminal.
func CanTransitionStore(from, to StoreStatus) bool {
	for _, candidate := range storeTransitions[from] {
		if candidate == to {
			return true
		}
	}
	return false
}

// StoreStats is the persisted read model used by lists and rankings. It is
// recomputed from review events and per-user summaries, never derived at read
// time.
type StoreStats struct {
	RatingCount      int
	RatingAvg        float64
	RatingScore      float64
	TasteAvg         float64
	ValueAvg         float64
	QueueAvg         float64
	HygieneAvg       float64
	ServiceAvg       float64
	ReviewEventCount int
	LastReviewAt     *time.Time
	RatingUpdatedAt  *time.Time
}

type Store struct {
	ID                int
	HallID            int
	Name              string
	NormalizedName    string
	Floor             string
	Categories        []string
	Status            StoreStatus
	MergedIntoStoreID *int
	CoverMediaID      *int
	Stats             StoreStats
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (s *Store) ApplyDefaults() {
	s.Name = strings.TrimSpace(s.Name)
	s.Floor = strings.TrimSpace(s.Floor)
	s.NormalizedName = NormalizeName(s.Name)
	s.Categories = NormalizeCategories(s.Categories)
	if s.Status == "" {
		s.Status = StoreStatusPending
	}
}

func (s *Store) Validate() error {
	if s.HallID <= 0 {
		return apperr.BadRequest("店铺必须归属食堂")
	}
	if s.Name == "" || s.NormalizedName == "" {
		return apperr.BadRequest("店铺名称不能为空")
	}
	if utf8.RuneCountInString(s.Name) > MaxStoreNameRunes {
		return apperr.BadRequest("店铺名称过长")
	}
	if utf8.RuneCountInString(s.Floor) > MaxFloorRunes {
		return apperr.BadRequest("楼层信息过长")
	}
	if len(s.Categories) > MaxStoreCategories {
		return apperr.BadRequest("店铺分类过多")
	}
	for _, category := range s.Categories {
		if utf8.RuneCountInString(category) > MaxCategoryRunes {
			return apperr.BadRequest("店铺分类过长")
		}
	}
	if !s.Status.Valid() {
		return apperr.BadRequest("店铺状态无效")
	}
	return nil
}

// CanAcceptReview guards new review events; offline stores keep their history
// readable but stop collecting new visits.
func (s *Store) CanAcceptReview() bool {
	return s.Status == StoreStatusActive
}

func (s *Store) PubliclyVisible() bool {
	return s.Status == StoreStatusActive || s.Status == StoreStatusOffline
}

func (s *Store) TransitionTo(to StoreStatus, now time.Time) error {
	if s.Status == to {
		return nil
	}
	if !CanTransitionStore(s.Status, to) {
		return apperr.BadRequest("店铺状态不允许该变更")
	}
	s.Status = to
	s.UpdatedAt = now
	return nil
}

type DishStatus string

const (
	DishStatusPending  DishStatus = "pending"
	DishStatusActive   DishStatus = "active"
	DishStatusHidden   DishStatus = "hidden"
	DishStatusRejected DishStatus = "rejected"
	DishStatusMerged   DishStatus = "merged"
)

func (s DishStatus) Valid() bool {
	switch s {
	case DishStatusPending, DishStatusActive, DishStatusHidden, DishStatusRejected, DishStatusMerged:
		return true
	default:
		return false
	}
}

var dishTransitions = map[DishStatus][]DishStatus{
	DishStatusPending:  {DishStatusActive, DishStatusRejected, DishStatusMerged},
	DishStatusRejected: {DishStatusActive, DishStatusMerged},
	DishStatusActive:   {DishStatusHidden, DishStatusMerged},
	DishStatusHidden:   {DishStatusActive, DishStatusMerged},
	DishStatusMerged:   {},
}

func CanTransitionDish(from, to DishStatus) bool {
	for _, candidate := range dishTransitions[from] {
		if candidate == to {
			return true
		}
	}
	return false
}

type DishStats struct {
	RatingCount      int
	RatingAvg        float64
	RatingScore      float64
	TasteAvg         float64
	PortionAvg       float64
	ValueAvg         float64
	ReviewEventCount int
	RatingUpdatedAt  *time.Time
}

type Dish struct {
	ID               int
	StoreID          int
	Name             string
	NormalizedName   string
	Status           DishStatus
	MergedIntoDishID *int
	Stats            DishStats
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (d *Dish) ApplyDefaults() {
	d.Name = strings.TrimSpace(d.Name)
	d.NormalizedName = NormalizeName(d.Name)
	if d.Status == "" {
		d.Status = DishStatusPending
	}
}

func (d *Dish) Validate() error {
	if d.StoreID <= 0 {
		return apperr.BadRequest("菜品必须归属店铺")
	}
	if d.Name == "" || d.NormalizedName == "" {
		return apperr.BadRequest("菜品名称不能为空")
	}
	if utf8.RuneCountInString(d.Name) > MaxDishNameRunes {
		return apperr.BadRequest("菜品名称过长")
	}
	if !d.Status.Valid() {
		return apperr.BadRequest("菜品状态无效")
	}
	return nil
}

func (d *Dish) PubliclyVisible() bool {
	return d.Status == DishStatusActive
}

func (d *Dish) TransitionTo(to DishStatus, now time.Time) error {
	if d.Status == to {
		return nil
	}
	if !CanTransitionDish(d.Status, to) {
		return apperr.BadRequest("菜品状态不允许该变更")
	}
	d.Status = to
	d.UpdatedAt = now
	return nil
}
