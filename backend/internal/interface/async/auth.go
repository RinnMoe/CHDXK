package async

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	domainauth "jcourse/internal/domain/auth"
)

type clearExpiredSuspensionHandler struct {
	authService *domainauth.AuthService
}

func newClearExpiredSuspensionHandler(authService *domainauth.AuthService) asynq.Handler {
	return &clearExpiredSuspensionHandler{authService: authService}
}

func (h *clearExpiredSuspensionHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload domainauth.ClearExpiredSuspensionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	if err := h.authService.ClearExpiredSuspension(ctx, payload.UserID); err != nil {
		return err
	}
	return nil
}
