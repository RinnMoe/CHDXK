package application

import (
	"context"
	"strings"

	"jcourse/internal/domain/setting"
	"jcourse/pkg/sitecopy"
	"jcourse/pkg/sitepages"
)

type SystemSettingsQueryService struct {
	settings *setting.SystemSettingsService
}

func (s *SystemSettingsQueryService) Pages(ctx context.Context) ([]sitepages.Page, error) {
	snapshot, err := s.settings.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	pages, err := sitepages.Parse(snapshot.String(sitepages.Key))
	if err != nil {
		return nil, err
	}
	result := make([]sitepages.Page, 0, len(pages))
	for _, page := range pages {
		if page.Published {
			result = append(result, page)
		}
	}
	return result, nil
}

func NewSystemSettingsQueryService(settings *setting.SystemSettingsService) *SystemSettingsQueryService {
	return &SystemSettingsQueryService{settings: settings}
}

func (s *SystemSettingsQueryService) List(ctx context.Context) ([]SystemSettingDTO, error) {
	settings, err := s.settings.List(ctx, false)
	if err != nil {
		return nil, err
	}
	return newSystemSettingDTOs(settings), nil
}

func (s *SystemSettingsQueryService) ListAdmin(ctx context.Context) ([]SystemSettingDTO, error) {
	settings, err := s.settings.List(ctx, true)
	if err != nil {
		return nil, err
	}
	return newSystemSettingDTOs(settings), nil
}

func newSystemSettingDTOs(settings []setting.EffectiveSystemSetting) []SystemSettingDTO {
	values := make(map[string]string)
	for _, item := range settings {
		if strings.HasPrefix(item.Definition.Key, "copy.") {
			values[item.Definition.Key] = item.Value
		}
	}
	ctx := sitecopy.WithValues(context.Background(), values)
	items := make([]SystemSettingDTO, len(settings))
	for i := range settings {
		items[i] = newSystemSettingDTO(settings[i])
		if !strings.HasPrefix(items[i].Key, "copy.") {
			items[i].Label = sitecopy.Text(ctx, items[i].Label)
			items[i].Description = sitecopy.Text(ctx, items[i].Description)
		}
	}
	return items
}

func (s *SystemSettingsQueryService) PublicCopy(ctx context.Context) (map[string]string, error) {
	items, err := s.settings.List(ctx, false)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string)
	for _, item := range items {
		if strings.HasPrefix(item.Definition.Key, "copy.") {
			values[item.Definition.Key] = item.Value
		}
	}
	return values, nil
}
