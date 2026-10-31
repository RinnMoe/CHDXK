package app

import (
	"context"

	"jcourse/internal/application"
	"jcourse/internal/domain/setting"
)

// systemCanteenSettingsProvider projects persisted system settings into the
// canteen runtime toggles without importing the setting registry into the
// domain.
type systemCanteenSettingsProvider struct {
	settings *setting.SystemSettingsService
}

func newSystemCanteenSettingsProvider(settings *setting.SystemSettingsService) *systemCanteenSettingsProvider {
	return &systemCanteenSettingsProvider{settings: settings}
}

func (p *systemCanteenSettingsProvider) CanteenRuntimeConfig(ctx context.Context) (application.CanteenRuntimeConfig, error) {
	snapshot, err := p.settings.Snapshot(ctx)
	if err != nil {
		return application.CanteenRuntimeConfig{}, err
	}
	return application.CanteenRuntimeConfig{
		Enabled:        snapshot.Bool(setting.SystemSettingKeyCanteenEnabled),
		ReviewsEnabled: snapshot.Bool(setting.SystemSettingKeyCanteenReviewsEnabled),
		MediaEnabled:   snapshot.Bool(setting.SystemSettingKeyCanteenMediaEnabled),
	}, nil
}

var _ application.CanteenSettingsProvider = (*systemCanteenSettingsProvider)(nil)
