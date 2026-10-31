package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"jcourse/internal/domain/canteen"
	"jcourse/pkg/apperr"
)

// CanteenReviewQueryService exposes review lists for the store page and the
// caller's own history.
type CanteenReviewQueryService struct {
	reviews  canteen.ReviewQuery
	settings CanteenSettingsProvider
}

func NewCanteenReviewQueryService(reviews canteen.ReviewQuery, settings CanteenSettingsProvider) *CanteenReviewQueryService {
	return &CanteenReviewQueryService{reviews: reviews, settings: settings}
}

func (s *CanteenReviewQueryService) requireEnabled(ctx context.Context) error {
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
	return nil
}

func (s *CanteenReviewQueryService) ListStoreReviews(
	ctx context.Context,
	storeID int,
	filter canteen.ReviewListFilter,
) (*CanteenReviewPage, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	filter.StoreID = storeID
	filter.Order = strings.ToLower(strings.TrimSpace(filter.Order))
	if filter.Order != "best" && filter.Order != "worst" {
		filter.Order = "latest"
	}
	filter.Page, filter.PageSize = normalizeCanteenPagination(filter.Page, filter.PageSize)
	if filter.Cursor != nil {
		if filter.Cursor.Order != filter.Order {
			return nil, apperr.BadRequest("评价分页游标无效或已改变排序")
		}
	}
	requested := filter.PageSize
	// Fetch one extra row to know whether another cursor page exists.
	filter.PageSize = requested + 1
	items, total, err := s.reviews.ListReviews(ctx, filter)
	if err != nil {
		return nil, err
	}
	nextCursor := ""
	if len(items) > requested {
		items = items[:requested]
		last := items[len(items)-1]
		nextCursor = EncodeCanteenReviewCursor(canteen.ReviewCursor{
			Order:        filter.Order,
			OverallScore: last.OverallScore,
			LikeCount:    last.LikeCount,
			CreatedAt:    last.CreatedAt,
			ID:           last.ID,
		})
	}
	page := newCanteenReviewPage(items, total, filter.Page, requested)
	page.NextCursor = nextCursor
	return page, nil
}

// EncodeCanteenReviewCursor serializes an opaque keyset position.
func EncodeCanteenReviewCursor(cursor canteen.ReviewCursor) string {
	raw, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeCanteenReviewCursor parses a caller-supplied keyset position.
func DecodeCanteenReviewCursor(raw string) (*canteen.ReviewCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return nil, apperr.BadRequest("评价分页游标无效")
	}
	var cursor canteen.ReviewCursor
	if err := json.Unmarshal(decoded, &cursor); err != nil || cursor.ID <= 0 || cursor.CreatedAt.IsZero() {
		return nil, apperr.BadRequest("评价分页游标无效")
	}
	return &cursor, nil
}

func (s *CanteenReviewQueryService) ListMyReviews(
	ctx context.Context,
	userID int,
	page int,
	pageSize int,
) (*CanteenReviewPage, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	page, pageSize = normalizeCanteenPagination(page, pageSize)
	items, total, err := s.reviews.ListUserReviews(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return newCanteenReviewPage(items, total, page, pageSize), nil
}

func newCanteenReviewPage(items []canteen.ReviewView, total int64, page, pageSize int) *CanteenReviewPage {
	list := make([]CanteenReviewDTO, 0, len(items))
	for _, item := range items {
		list = append(list, newCanteenReviewDTO(item))
	}
	return &CanteenReviewPage{Items: list, Total: total, Page: page, PageSize: pageSize}
}
