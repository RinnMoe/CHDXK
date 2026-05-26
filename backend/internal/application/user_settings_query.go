package application

import (
	"context"
	"strings"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
)

type UserSettingsQueryService struct {
	courseRepo course.CourseRepository
	settings   *setting.UserSettingsService
}

func NewUserSettingsQueryService(repo setting.Repository, courseRepo course.CourseRepository) *UserSettingsQueryService {
	return &UserSettingsQueryService{courseRepo: courseRepo, settings: setting.NewUserSettingsService(repo, courseRepo)}
}

func (s *UserSettingsQueryService) Get(ctx context.Context, userID int) (*UserSettingsDTO, error) {
	semesters, err := s.validSemesters(ctx)
	if err != nil {
		return nil, err
	}
	current, err := s.settings.GetCurrentSemester(ctx, userID, semesters)
	if err != nil {
		return nil, err
	}
	return &UserSettingsDTO{CurrentSemester: current}, nil
}

func (s *UserSettingsQueryService) validSemesters(ctx context.Context) ([]string, error) {
	return validSemesters(ctx, s.courseRepo)
}

func validSemesters(ctx context.Context, courseRepo course.CourseRepository) ([]string, error) {
	filters, err := courseRepo.GetFilters(ctx)
	if err != nil {
		return nil, err
	}
	if filters == nil {
		return []string{}, nil
	}
	semesters := make([]string, 0, len(filters.Semesters))
	for _, item := range filters.Semesters {
		name := strings.TrimSpace(item.Name)
		if name != "" {
			semesters = append(semesters, name)
		}
	}
	return semesters, nil
}
