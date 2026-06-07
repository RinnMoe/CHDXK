//go:build test

package setting

import "context"

type MockSystemRepository struct {
	Settings map[string]*SystemSetting

	OnList func(context.Context) ([]SystemSetting, error)
	OnGet  func(context.Context, string) (*SystemSetting, error)
	OnSave func(context.Context, *SystemSetting) error
}

func NewMockSystemRepository() *MockSystemRepository {
	return &MockSystemRepository{Settings: map[string]*SystemSetting{}}
}

func (r *MockSystemRepository) List(ctx context.Context) ([]SystemSetting, error) {
	if r.OnList != nil {
		return r.OnList(ctx)
	}
	r.ensureMap()
	settings := make([]SystemSetting, 0, len(r.Settings))
	for _, item := range r.Settings {
		settings = append(settings, *item)
	}
	return settings, nil
}

func (r *MockSystemRepository) Get(ctx context.Context, key string) (*SystemSetting, error) {
	if r.OnGet != nil {
		return r.OnGet(ctx, key)
	}
	r.ensureMap()
	setting, ok := r.Settings[key]
	if !ok {
		return nil, nil
	}
	copy := *setting
	return &copy, nil
}

func (r *MockSystemRepository) Save(ctx context.Context, setting *SystemSetting) error {
	if r.OnSave != nil {
		return r.OnSave(ctx, setting)
	}
	r.ensureMap()
	copy := *setting
	r.Settings[setting.Key] = &copy
	return nil
}

func (r *MockSystemRepository) ensureMap() {
	if r.Settings == nil {
		r.Settings = map[string]*SystemSetting{}
	}
}
