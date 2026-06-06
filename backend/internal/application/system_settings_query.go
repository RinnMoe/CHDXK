package application

import (
	"context"

	"jcourse/internal/domain/setting"
)

type SystemSettingsQueryService struct {
	settings *setting.SystemSettingsService
}

func NewSystemSettingsQueryService(repo setting.SystemRepository) *SystemSettingsQueryService {
	return &SystemSettingsQueryService{settings: setting.NewSystemSettingsService(repo)}
}

func (s *SystemSettingsQueryService) List(ctx context.Context) ([]SystemSettingDTO, error) {
	settings, err := s.settings.List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]SystemSettingDTO, len(settings))
	for i := range settings {
		items[i] = newSystemSettingDTO(settings[i])
	}
	return items, nil
}
