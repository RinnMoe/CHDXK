package canteen

import (
	"strings"
	"time"
	"unicode/utf8"

	"jcourse/pkg/apperr"
)

const (
	MaxDishReviewContentRunes = 500
	MaxDishSuggestions        = 8
)

// DishReviewEvent is one independent dish rating. It may optionally link to a
// visit review, but it never changes store-level aggregates.
type DishReviewEvent struct {
	ID            int
	DishID        int
	UserID        int
	ReviewEventID *int
	DishDimensions
	OverallScore float64
	Content      string
	Status       ReviewStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r *DishReviewEvent) ApplyCreate(weights DishWeights) {
	r.Content = strings.TrimSpace(r.Content)
	r.OverallScore = weights.Score(r.DishDimensions)
	if r.Status == "" {
		r.Status = ReviewStatusActive
	}
}

func (r *DishReviewEvent) Validate() error {
	if r.DishID <= 0 || r.UserID <= 0 {
		return apperr.BadRequest("菜品评价归属无效")
	}
	if !r.Status.Valid() {
		return apperr.BadRequest("菜品评价状态无效")
	}
	if err := r.DishDimensions.Validate(); err != nil {
		return err
	}
	if utf8.RuneCountInString(r.Content) > MaxDishReviewContentRunes {
		return apperr.BadRequest("菜品评价不能超过 500 个字符")
	}
	return nil
}

// DishRatingSummary is one user's weighted sample for a dish.
type DishRatingSummary struct {
	DishID             int
	UserID             int
	WeightedScore      float64
	WindowSize         int
	LifetimeEventCount int
	LastReviewedAt     *time.Time
	ComputedAt         time.Time
}

// DishReviewView is the read model for a user's dish rating.
type DishReviewView struct {
	ID           int
	DishID       int
	DishName     string
	StoreID      int
	StoreName    string
	UserID       int
	Taste        int
	Portion      int
	Value        int
	OverallScore float64
	Content      string
	Status       ReviewStatus
	CreatedAt    time.Time
}

// MediaFilter drives the admin media queue.
type MediaFilter struct {
	Status   MediaStatus
	Page     int
	PageSize int
}
