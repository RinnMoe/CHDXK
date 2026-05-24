package application

import (
	"context"
	"strings"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
)

type UserSettingsDTO struct {
	CurrentSemester string `json:"current_semester"`
}

type UpdateUserSettingsCommand struct {
	CurrentSemester string `json:"current_semester"`
}

type UserSettingsQueryService struct {
	repo        setting.Repository
	courseQuery course.CourseQuery
	courseRepo  course.CourseRepository
}

func NewUserSettingsQueryService(repo setting.Repository, courseQuery course.CourseQuery, courseRepo course.CourseRepository) *UserSettingsQueryService {
	return &UserSettingsQueryService{repo: repo, courseQuery: courseQuery, courseRepo: courseRepo}
}

func (s *UserSettingsQueryService) Get(ctx context.Context, userID int) (*UserSettingsDTO, error) {
	settings, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	semesters, err := s.validSemesters(ctx)
	if err != nil {
		return nil, err
	}

	current := ""
	if len(semesters) > 0 {
		current = semesters[0]
	}
	if settings != nil && settings.CurrentSemester != "" {
		allowed, err := s.courseRepo.OfferedSemesterExists(ctx, settings.CurrentSemester)
		if err != nil {
			return nil, err
		}
		if allowed {
			current = settings.CurrentSemester
		}
	}
	return &UserSettingsDTO{CurrentSemester: current}, nil
}

type UserSettingsCommandService struct {
	repo       setting.Repository
	courseRepo course.CourseRepository
}

func NewUserSettingsCommandService(repo setting.Repository, courseRepo course.CourseRepository) *UserSettingsCommandService {
	return &UserSettingsCommandService{repo: repo, courseRepo: courseRepo}
}

func (s *UserSettingsCommandService) Update(ctx context.Context, userID int, cmd UpdateUserSettingsCommand) (*UserSettingsDTO, error) {
	currentSemester := strings.TrimSpace(cmd.CurrentSemester)
	if currentSemester == "" {
		return nil, setting.ErrInvalidCurrentSemester
	}
	allowed, err := s.courseRepo.OfferedSemesterExists(ctx, currentSemester)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, setting.ErrInvalidCurrentSemester
	}

	settings := &setting.UserSettings{UserID: userID, CurrentSemester: currentSemester}
	if err := s.repo.Save(ctx, settings); err != nil {
		return nil, err
	}
	return &UserSettingsDTO{CurrentSemester: settings.CurrentSemester}, nil
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
