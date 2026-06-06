package application

import (
	"context"
	"strings"
	"time"

	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
)

type SystemSettingsCommandService struct {
	settings *setting.SystemSettingsService
}

func NewSystemSettingsCommandService(repo setting.SystemRepository, courseRepo course.CourseRepository) *SystemSettingsCommandService {
	validators := setting.NewValueValidatorFactory(map[string]setting.ValueValidator{
		setting.SystemSettingKeyCurrentSemester: currentSemesterValidator{courseRepo: courseRepo},
	})
	return &SystemSettingsCommandService{settings: setting.NewSystemSettingsService(repo, validators)}
}

func (s *SystemSettingsCommandService) Update(ctx context.Context, actor *auth.User, key string, cmd UpdateSystemSettingCommand) (*SystemSettingDTO, error) {
	previousSetting, err := s.settings.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	updatedSetting, err := s.settings.Save(ctx, key, cmd.Value)
	if err != nil {
		return nil, err
	}

	previousValue := ""
	if previousSetting != nil {
		previousValue = previousSetting.Value
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  time.Now(),
		ActorUserID: actor.ID,
		Action:      audit.ActionSystemSettingsUpdate,
		TargetType:  audit.TargetTypeSystemSettings,
		TargetID:    updatedSetting.Key,
		Details: audit.Details{
			"from": previousValue,
			"to":   updatedSetting.Value,
		},
	})

	dto := newSystemSettingDTO(*updatedSetting)
	return &dto, nil
}

type currentSemesterValidator struct {
	courseRepo course.CourseRepository
}

func (v currentSemesterValidator) Validate(ctx context.Context, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return setting.ErrInvalidCurrentSemester
	}
	allowed, err := v.courseRepo.OfferedSemesterExists(ctx, value)
	if err != nil {
		return err
	}
	if !allowed {
		return setting.ErrInvalidCurrentSemester
	}
	return nil
}
