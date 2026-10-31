package application

import "context"

// CanteenRuntimeConfig exposes the operational toggles for the canteen domain.
type CanteenRuntimeConfig struct {
	Enabled        bool
	ReviewsEnabled bool
	MediaEnabled   bool
}

// CanteenSettingsProvider keeps the canteen domain independent of the system
// settings registry while still allowing runtime toggles.
type CanteenSettingsProvider interface {
	CanteenRuntimeConfig(ctx context.Context) (CanteenRuntimeConfig, error)
}

// StaticCanteenSettingsProvider is used by tests and by wiring that does not
// depend on persisted settings.
type StaticCanteenSettingsProvider struct {
	Runtime CanteenRuntimeConfig
}

func (p StaticCanteenSettingsProvider) CanteenRuntimeConfig(context.Context) (CanteenRuntimeConfig, error) {
	return p.Runtime, nil
}

var _ CanteenSettingsProvider = StaticCanteenSettingsProvider{}
