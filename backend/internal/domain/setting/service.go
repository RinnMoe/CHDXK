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

func NewUserSettingsService(repo Repository, courseRepo course.CourseRepository) *UserSettingsService {
	return &UserSettingsService{repo: repo, courseRepo: courseRepo}
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
