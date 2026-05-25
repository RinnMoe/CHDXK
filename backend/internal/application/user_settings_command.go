package application

import (
	"context"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
)

type UserSettingsCommandService struct {
	settings *setting.UserSettingsService
}

func NewUserSettingsCommandService(repo setting.Repository, courseRepo course.CourseRepository) *UserSettingsCommandService {
	return &UserSettingsCommandService{settings: setting.NewUserSettingsService(repo, courseRepo)}
}

func (s *UserSettingsCommandService) Update(ctx context.Context, userID int, cmd UpdateUserSettingsCommand) (*UserSettingsDTO, error) {
	settings, err := s.settings.UpdateCurrentSemester(ctx, userID, cmd.CurrentSemester)
	if err != nil {
		return nil, err
	}
	return &UserSettingsDTO{CurrentSemester: settings.CurrentSemester}, nil
}
