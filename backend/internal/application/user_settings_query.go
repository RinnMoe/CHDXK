package application

import (
	"context"
	"strings"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
)

type UserSettingsQueryService struct {
	courseQuery course.CourseQuery
	settings    *setting.UserSettingsService
}

func NewUserSettingsQueryService(repo setting.Repository, courseQuery course.CourseQuery, courseRepo course.CourseRepository) *UserSettingsQueryService {
	return &UserSettingsQueryService{courseQuery: courseQuery, settings: setting.NewUserSettingsService(repo, courseRepo)}
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
	return validSemesters(ctx, s.courseQuery)
}

func validSemesters(ctx context.Context, courseQuery course.CourseQuery) ([]string, error) {
	filters, err := courseQuery.GetFilters(ctx)
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
