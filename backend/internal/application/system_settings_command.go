package application

import (
	"context"
	"time"

	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/setting"
)

type SystemSettingsCommandService struct {
	settings *setting.SystemSettingsService
}

func NewSystemSettingsCommandService(settings *setting.SystemSettingsService) *SystemSettingsCommandService {
	return &SystemSettingsCommandService{settings: settings}
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

	def, _ := s.settings.Definition(updatedSetting.Key)
	return new(newSystemSettingDTO(setting.EffectiveSystemSetting{
		Definition: def,
		Value:      updatedSetting.Value,
	})), nil
}
