package application

import (
	"context"
	"strconv"
	"time"

	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
)

type ApiKeyCommandService struct {
	svc      *auth.ApiKeyService
	settings SiteSettingsProvider
}

func NewApiKeyCommandService(svc *auth.ApiKeyService, settings SiteSettingsProvider) *ApiKeyCommandService {
	if settings == nil {
		defaults := NewDefaultSiteSettingsProvider()
		settings = defaults
	}
	return &ApiKeyCommandService{svc: svc, settings: settings}
}

func (s *ApiKeyCommandService) CreateMyApiKey(ctx context.Context, userID int, cmd CreateApiKeyCommand) (*ApiKeyDTO, error) {
	config, err := s.settings.ApiKeyConfig(ctx)
	if err != nil {
		return nil, err
	}
	key, credential, err := s.svc.CreateUserKeyWithConfig(ctx, userID, cmd.Name, config)
	if err != nil {
		return nil, err
	}
	dto := newApiKeyDTO(*key, credential.Key())
	return &dto, nil
}

func (s *ApiKeyCommandService) DeleteMyApiKey(ctx context.Context, userID int, id int64) error {
	return s.svc.DeleteUserKey(ctx, userID, id)
}

func (s *ApiKeyCommandService) CreateSystemApiKey(ctx context.Context, actor *auth.User, cmd CreateApiKeyCommand) (*ApiKeyDTO, error) {
	key, credential, err := s.svc.CreateSystemKey(ctx, cmd.Name)
	if err != nil {
		return nil, err
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  time.Now(),
		ActorUserID: actor.ID,
		Action:      audit.ActionSystemAPIKeyCreate,
		TargetType:  audit.TargetTypeSystemAPIKey,
		TargetID:    strconv.FormatInt(key.ID, 10),
		Details: audit.Details{
			"name": key.Name,
		},
	})
	dto := newApiKeyDTO(*key, credential.Key())
	return &dto, nil
}

func (s *ApiKeyCommandService) DeleteSystemApiKey(ctx context.Context, actor *auth.User, id int64) error {
	key, err := s.svc.GetKey(ctx, id)
	if err != nil {
		return err
	}
	err = s.svc.DeleteSystemKey(ctx, id)
	if err != nil {
		return err
	}
	details := audit.Details{}
	if key != nil {
		details["name"] = key.Name
	}
	audit.EnqueueLog(ctx, audit.Log{
		OccurredAt:  time.Now(),
		ActorUserID: actor.ID,
		Action:      audit.ActionSystemAPIKeyDelete,
		TargetType:  audit.TargetTypeSystemAPIKey,
		TargetID:    strconv.FormatInt(id, 10),
		Details:     details,
	})
	return nil
}
