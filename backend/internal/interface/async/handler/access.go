package handler

import (
	"context"

	"github.com/hibiken/asynq"

	"jcourse/internal/domain/auth"
)

type flushAccessHandler struct {
	tracker auth.AccessTracker
}

func NewFlushAccessHandler(tracker auth.AccessTracker) asynq.Handler {
	return &flushAccessHandler{tracker: tracker}
}

func (h *flushAccessHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	_, err := h.tracker.Flush(ctx)
	return err
}
