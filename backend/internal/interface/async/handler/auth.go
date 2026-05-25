package handler

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	domainauth "jcourse/internal/domain/auth"
)

type clearExpiredSuspensionHandler struct {
	currentUserService *domainauth.AuthUserService
}

type suspendUserHandler struct {
	currentUserService *domainauth.AuthUserService
}

func NewClearExpiredSuspensionHandler(currentUserService *domainauth.AuthUserService) asynq.Handler {
	return &clearExpiredSuspensionHandler{currentUserService: currentUserService}
}

func NewSuspendUserHandler(currentUserService *domainauth.AuthUserService) asynq.Handler {
	return &suspendUserHandler{currentUserService: currentUserService}
}

func (h *clearExpiredSuspensionHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload domainauth.ClearExpiredSuspensionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	if err := h.currentUserService.ClearExpiredSuspension(ctx, payload.UserID); err != nil {
		return err
	}
	return nil
}

func (h *suspendUserHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload domainauth.SuspendUserPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	return h.currentUserService.SuspendUser(ctx, payload.UserID, payload.Duration)
}
