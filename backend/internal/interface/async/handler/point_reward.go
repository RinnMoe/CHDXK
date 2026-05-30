package handler

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"jcourse/internal/application"
	"jcourse/internal/domain/point"
)

type grantPointRewardHandler struct {
	command *application.PointRewardCommandService
}

func NewGrantPointRewardHandler(command *application.PointRewardCommandService) asynq.Handler {
	return &grantPointRewardHandler{command: command}
}

func (h *grantPointRewardHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload point.GrantRewardPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	return h.command.GrantReward(ctx, payload.RewardID)
}
