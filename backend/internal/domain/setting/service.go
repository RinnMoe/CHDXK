package setting

import (
	"context"
	"strings"
)

type SystemSettingsService struct {
	repo       SystemRepository
	validators *ValueValidatorFactory
}

func NewSystemSettingsService(repo SystemRepository, validators ...*ValueValidatorFactory) *SystemSettingsService {
	var factory *ValueValidatorFactory
	if len(validators) > 0 {
		factory = validators[0]
	}
	return &SystemSettingsService{repo: repo, validators: factory}
}

func (s *SystemSettingsService) List(ctx context.Context) ([]SystemSetting, error) {
	return s.repo.List(ctx)
}

func (s *SystemSettingsService) Get(ctx context.Context, key string) (*SystemSetting, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidSystemSettingKey
	}
	return s.repo.Get(ctx, key)
}

func (s *SystemSettingsService) Save(ctx context.Context, key, value string) (*SystemSetting, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidSystemSettingKey
	}
	if !IsRegisteredSystemSettingKey(key) {
		return nil, ErrInvalidSystemSettingKey
	}
	if err := s.validators.Validate(ctx, key, value); err != nil {
		return nil, err
	}
	setting := &SystemSetting{Key: key, Value: value}
	if err := s.repo.Save(ctx, setting); err != nil {
		return nil, err
	}
	return setting, nil
}
