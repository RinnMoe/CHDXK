package application

import (
	"context"
	"errors"
	"strconv"
	"time"

	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

type CreateCanteenDishCommand struct {
	StoreID int
	Name    string
}

type CreateCanteenDishReviewCommand struct {
	DishID        int
	ReviewEventID *int
	Taste         int
	Portion       int
	Value         int
	Content       string
	MediaIDs      []int
}

// CanteenDishCommandService maintains the dish directory and dish ratings.
type CanteenDishCommandService struct {
	stores      canteen.StoreRepository
	dishes      canteen.DishRepository
	dishQuery   canteen.DishQuery
	dishReviews canteen.DishReviewRepository
	settings    CanteenSettingsProvider
	media       *CanteenMediaCommandService
	now         func() time.Time
}

func NewCanteenDishCommandService(
	stores canteen.StoreRepository,
	dishes canteen.DishRepository,
	dishQuery canteen.DishQuery,
	dishReviews canteen.DishReviewRepository,
	settings CanteenSettingsProvider,
	media *CanteenMediaCommandService,
) *CanteenDishCommandService {
	return &CanteenDishCommandService{
		stores:      stores,
		dishes:      dishes,
		dishQuery:   dishQuery,
		dishReviews: dishReviews,
		settings:    settings,
		media:       media,
		now:         time.Now,
	}
}

func (s *CanteenDishCommandService) requireWritable(ctx context.Context) error {
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
	if !config.ReviewsEnabled {
		return apperr.Forbidden("到店评价暂未开放")
	}
	return nil
}

// CreateDish lets users contribute a pending dish and administrators create an
// active one. Duplicate normalized names return the existing dish.
func (s *CanteenDishCommandService) CreateDish(ctx context.Context, user *auth.User, cmd CreateCanteenDishCommand) (*canteen.Dish, bool, error) {
	if s.settings != nil {
		config, err := s.settings.CanteenRuntimeConfig(ctx)
		if err != nil {
			return nil, false, err
		}
		if !config.Enabled {
			return nil, false, apperr.NotFound("食堂评价暂未开放")
		}
	}
	store, err := s.stores.Get(ctx, cmd.StoreID)
	if err != nil {
		return nil, false, err
	}
	if store == nil {
		return nil, false, canteen.ErrStoreNotFound
	}
	if !store.PubliclyVisible() {
		return nil, false, canteen.ErrStoreNotFound
	}
	normalized := canteen.NormalizeName(cmd.Name)
	if normalized == "" {
		return nil, false, apperr.BadRequest("菜品名称不能为空")
	}
	existing, err := s.dishes.GetByStoreAndNormalizedName(ctx, cmd.StoreID, normalized)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, true, nil
	}
	status := canteen.DishStatusPending
	if user != nil && user.IsAdmin() {
		status = canteen.DishStatusActive
	}
	now := s.now()
	dish := &canteen.Dish{
		StoreID:   cmd.StoreID,
		Name:      cmd.Name,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	dish.ApplyDefaults()
	if err := dish.Validate(); err != nil {
		return nil, false, err
	}
	if err := s.dishes.Create(ctx, dish); err != nil {
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr.StatusCode == 409 {
			existing, findErr := s.dishes.GetByStoreAndNormalizedName(ctx, cmd.StoreID, normalized)
			if findErr != nil {
				return nil, false, findErr
			}
			if existing != nil {
				return existing, true, nil
			}
		}
		return nil, false, err
	}
	if user != nil {
		audit.EnqueueLog(ctx, audit.Log{
			OccurredAt:  now,
			ActorUserID: user.ID,
			Action:      "canteen.dish.create",
			TargetType:  "canteen_dish",
			TargetID:    strconv.Itoa(dish.ID),
			Details:     audit.Details{"store_id": dish.StoreID, "name": dish.Name},
		})
	}
	return dish, false, nil
}

func (s *CanteenDishCommandService) UpdateDishStatus(ctx context.Context, user *auth.User, dishID int, status canteen.DishStatus) (*canteen.Dish, error) {
	if user == nil || !user.IsAdmin() {
		return nil, canteen.ErrCannotModerate
	}
	dish, err := s.dishes.Get(ctx, dishID)
	if err != nil {
		return nil, err
	}
	if dish == nil {
		return nil, canteen.ErrDishNotFound
	}
	now := s.now()
	if err := dish.TransitionTo(status, now); err != nil {
		return nil, err
	}
	if err := s.dishes.UpdateStatus(ctx, dish.ID, status, now); err != nil {
		return nil, err
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  now,
		ActorUserID: user.ID,
		Action:      "canteen.dish.status",
		TargetType:  "canteen_dish",
		TargetID:    strconv.Itoa(dish.ID),
		Details:     audit.Details{"status": string(status)},
	})
	return dish, nil
}

func (s *CanteenDishCommandService) MergeDish(ctx context.Context, user *auth.User, dishID, targetID int) error {
	if user == nil || !user.IsAdmin() {
		return canteen.ErrCannotModerate
	}
	if dishID == targetID {
		return apperr.BadRequest("不能合并到自身")
	}
	dish, err := s.dishes.Get(ctx, dishID)
	if err != nil {
		return err
	}
	if dish == nil {
		return canteen.ErrDishNotFound
	}
	if dish.Status == canteen.DishStatusMerged {
		return canteen.ErrDishMerged
	}
	target, err := s.dishes.Get(ctx, targetID)
	if err != nil {
		return err
	}
	if target == nil || target.StoreID != dish.StoreID {
		return canteen.ErrDishNotFound
	}
	now := s.now()
	if err := s.dishes.Merge(ctx, dish.ID, target.ID, now); err != nil {
		return err
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  now,
		ActorUserID: user.ID,
		Action:      "canteen.dish.merge",
		TargetType:  "canteen_dish",
		TargetID:    strconv.Itoa(dish.ID),
		Details:     audit.Details{"merged_into": target.ID},
	})
	return nil
}

func (s *CanteenDishCommandService) CreateDishReview(ctx context.Context, user *auth.User, cmd CreateCanteenDishReviewCommand) (*canteen.DishReviewEvent, error) {
	if err := s.requireWritable(ctx); err != nil {
		return nil, err
	}
	if !canteen.CanCreateReview(user) {
		if user != nil && user.IsSuspended() {
			return nil, apperr.ErrUserSuspended
		}
		return nil, auth.ErrMuted
	}
	dish, err := s.dishQuery.GetDish(ctx, cmd.DishID)
	if err != nil {
		return nil, err
	}
	if dish == nil || dish.Status != canteen.DishStatusActive {
		return nil, canteen.ErrDishNotFound
	}
	store, err := s.stores.Get(ctx, dish.StoreID)
	if err != nil {
		return nil, err
	}
	if store == nil || !store.CanAcceptReview() {
		return nil, canteen.ErrStoreNotActive
	}
	now := s.now()
	review := &canteen.DishReviewEvent{
		DishID:        cmd.DishID,
		UserID:        user.ID,
		ReviewEventID: cmd.ReviewEventID,
		DishDimensions: canteen.DishDimensions{
			Taste:   cmd.Taste,
			Portion: cmd.Portion,
			Value:   cmd.Value,
		},
		Content:   cmd.Content,
		Status:    canteen.ReviewStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	review.ApplyCreate(canteen.DefaultDishWeights)
	if err := review.Validate(); err != nil {
		return nil, err
	}
	if s.media != nil && len(cmd.MediaIDs) > 0 {
		if _, err := s.media.ValidateAttachments(ctx, user, cmd.MediaIDs); err != nil {
			return nil, err
		}
	}
	if err := s.dishReviews.Create(ctx, review); err != nil {
		return nil, err
	}
	if s.media != nil && len(cmd.MediaIDs) > 0 {
		if err := s.media.AttachToDish(ctx, user, review.DishID, cmd.MediaIDs); err != nil {
			return nil, err
		}
	}
	return review, nil
}
