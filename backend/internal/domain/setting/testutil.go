//go:build test

package setting

import "context"

type MockRepository struct {
	Settings map[int]*UserSettings

	OnGetByUserID func(context.Context, int) (*UserSettings, error)
	OnSave        func(context.Context, *UserSettings) error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{Settings: map[int]*UserSettings{}}
}

func (r *MockRepository) GetByUserID(ctx context.Context, userID int) (*UserSettings, error) {
	if r.OnGetByUserID != nil {
		return r.OnGetByUserID(ctx, userID)
	}
	r.ensureMap()
	settings, ok := r.Settings[userID]
	if !ok {
		return nil, nil
	}
	copy := *settings
	return &copy, nil
}

func (r *MockRepository) Save(ctx context.Context, settings *UserSettings) error {
	if r.OnSave != nil {
		return r.OnSave(ctx, settings)
	}
	r.ensureMap()
	copy := *settings
	r.Settings[settings.UserID] = &copy
	return nil
}

func (r *MockRepository) ensureMap() {
	if r.Settings == nil {
		r.Settings = map[int]*UserSettings{}
	}
}
