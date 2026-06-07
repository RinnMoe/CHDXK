package application

import (
	"context"

	"jcourse/internal/domain/setting"
)

type SystemSettingsQueryService struct {
	settings *setting.SystemSettingsService
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
	items := make([]SystemSettingDTO, len(settings))
	for i := range settings {
		items[i] = newSystemSettingDTO(settings[i])
	}
	return items
}
