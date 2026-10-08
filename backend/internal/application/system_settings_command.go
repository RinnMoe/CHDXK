package application

import (
	"context"
	"jcourse/pkg/apperr"
	"strconv"
	"strings"
	"sync"
	"time"

	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/setting"
)

type SystemSettingsCommandService struct {
	settings       *setting.SystemSettingsService
	collectorQueue *CollectorQueue
	importMu       sync.Mutex
}

func NewSystemSettingsCommandService(settings *setting.SystemSettingsService) *SystemSettingsCommandService {
	return &SystemSettingsCommandService{settings: settings}
}

func (s *SystemSettingsCommandService) SetCollectorQueue(q *CollectorQueue) { s.collectorQueue = q }

func (s *SystemSettingsCommandService) Update(ctx context.Context, actor *auth.User, key string, cmd UpdateSystemSettingCommand) (*SystemSettingDTO, error) {
	if strings.HasPrefix(strings.TrimSpace(key), "import.") && (actor == nil || !actor.IsSuperAdmin()) {
		return nil, apperr.Forbidden("仅超级管理员可修改导入设置")
	}
	if strings.HasPrefix(strings.TrimSpace(key), "import.") {
		s.importMu.Lock()
		defer s.importMu.Unlock()
	}
	previousSetting, err := s.settings.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	updatedSetting, err := s.settings.Save(ctx, key, cmd.Value)
	if err != nil {
		return nil, err
	}

	if updatedSetting.Key == ImportConcurrency && s.collectorQueue != nil {
		n, _ := strconv.Atoi(updatedSetting.Value)
		s.collectorQueue.SetConcurrency(n)
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
