package application

import (
	"context"

	"jcourse/internal/domain/canteen"
)

// CanteenMediaQueryService lists media metadata for owners and moderators.
type CanteenMediaQueryService struct {
	query    canteen.MediaQuery
	settings CanteenSettingsProvider
}

func NewCanteenMediaQueryService(query canteen.MediaQuery, settings CanteenSettingsProvider) *CanteenMediaQueryService {
	return &CanteenMediaQueryService{query: query, settings: settings}
}

func (s *CanteenMediaQueryService) requireEnabled(ctx context.Context) error {
	if s.settings == nil {
		return nil
	}
	config, err := s.settings.CanteenRuntimeConfig(ctx)
	if err != nil {
		return err
	}
	if !config.Enabled {
		return canteen.ErrMediaNotFound
	}
	return nil
}

func (s *CanteenMediaQueryService) ListOwnerAssets(ctx context.Context, userID int, page, pageSize int) ([]CanteenMediaAssetDTO, int64, int, int, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, 0, 0, 0, err
	}
	page, pageSize = normalizeCanteenPagination(page, pageSize)
	items, total, err := s.query.ListMediaByOwner(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	result := make([]CanteenMediaAssetDTO, 0, len(items))
	for _, item := range items {
		result = append(result, newCanteenMediaAssetDTO(item))
	}
	return result, total, page, pageSize, nil
}

func (s *CanteenMediaQueryService) ListAdminAssets(ctx context.Context, filter canteen.MediaFilter) ([]CanteenAdminMediaDTO, int64, int, int, error) {
	page, pageSize := normalizeCanteenPagination(filter.Page, filter.PageSize)
	filter.Page, filter.PageSize = page, pageSize
	items, total, err := s.query.ListMediaByStatus(ctx, filter)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	result := make([]CanteenAdminMediaDTO, 0, len(items))
	for _, item := range items {
		result = append(result, newCanteenAdminMediaDTO(item))
	}
	return result, total, page, pageSize, nil
}
