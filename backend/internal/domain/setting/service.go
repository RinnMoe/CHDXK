package setting

import (
	"context"
	"strings"

	"jcourse/internal/domain/course"
)

type UserSettingsService struct {
	repo       Repository
	courseRepo course.CourseRepository
}

type SystemSettingsService struct {
	repo       SystemRepository
	validators *ValueValidatorFactory
}

func NewUserSettingsService(repo Repository, courseRepo course.CourseRepository) *UserSettingsService {
	return &UserSettingsService{repo: repo, courseRepo: courseRepo}
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

func (s *UserSettingsService) GetCurrentSemester(ctx context.Context, userID int, defaultSemesters []string) (string, error) {
	settings, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	current := firstNonEmpty(defaultSemesters)
	if settings != nil && strings.TrimSpace(settings.CurrentSemester) != "" {
		allowed, err := s.courseRepo.OfferedSemesterExists(ctx, settings.CurrentSemester)
		if err != nil {
			return "", err
		}
		if allowed {
			current = settings.CurrentSemester
		}
	}
	return current, nil
}

func (s *UserSettingsService) UpdateCurrentSemester(ctx context.Context, userID int, currentSemester string) (*UserSettings, error) {
	currentSemester = strings.TrimSpace(currentSemester)
	if currentSemester == "" {
		return nil, ErrInvalidCurrentSemester
	}
	allowed, err := s.courseRepo.OfferedSemesterExists(ctx, currentSemester)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrInvalidCurrentSemester
	}

	settings := &UserSettings{UserID: userID, CurrentSemester: currentSemester}
	if err := s.repo.Save(ctx, settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func firstNonEmpty(values []string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
