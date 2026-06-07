package setting

import (
	"context"
	"strings"
)

type SystemSettingsService struct {
	repo     SystemRepository
	registry *Registry
}

func NewSystemSettingsService(repo SystemRepository, registry *Registry) *SystemSettingsService {
	return &SystemSettingsService{repo: repo, registry: registry}
}

func (s *SystemSettingsService) List(ctx context.Context, includePrivate bool) ([]EffectiveSystemSetting, error) {
	defs := s.registry.List(includePrivate)
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	values := make(map[string]string, len(rows))
	for _, item := range rows {
		values[item.Key] = item.Value
	}

	settings := make([]EffectiveSystemSetting, 0, len(defs))
	for _, def := range defs {
		value := def.DefaultValue
		if override, ok := values[def.Key]; ok {
			value = override
		}
		settings = append(settings, EffectiveSystemSetting{
			Definition: def,
			Value:      value,
		})
	}
	return settings, nil
}

func (s *SystemSettingsService) Get(ctx context.Context, key string) (*SystemSetting, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidSystemSettingKey
	}
	return s.repo.Get(ctx, key)
}

func (s *SystemSettingsService) Definition(key string) (Definition, bool) {
	return s.registry.Get(strings.TrimSpace(key))
}

func (s *SystemSettingsService) Snapshot(ctx context.Context) (*Snapshot, error) {
	items, err := s.List(ctx, true)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string, len(items))
	for _, item := range items {
		values[item.Definition.Key] = item.Value
	}
	return NewSnapshot(values), nil
}

func (s *SystemSettingsService) Save(ctx context.Context, key, value string) (*SystemSetting, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidSystemSettingKey
	}
	def, ok := s.registry.Get(key)
	if !ok {
		return nil, ErrInvalidSystemSettingKey
	}
	if err := def.Validate(ctx, value); err != nil {
		return nil, err
	}
	setting := &SystemSetting{Key: key, Value: def.Normalize(value)}
	if err := s.repo.Save(ctx, setting); err != nil {
		return nil, err
	}
	return setting, nil
}
