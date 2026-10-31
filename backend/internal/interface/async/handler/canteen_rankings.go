package handler

import (
	"context"

	"github.com/hibiken/asynq"

	"jcourse/internal/application"
)

type refreshCanteenRankingsHandler struct {
	command *application.CanteenReviewCommandService
}

func NewRefreshCanteenRankingsHandler(command *application.CanteenReviewCommandService) asynq.Handler {
	return &refreshCanteenRankingsHandler{command: command}
}

func (h *refreshCanteenRankingsHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	return h.command.RefreshRankings(ctx)
}
