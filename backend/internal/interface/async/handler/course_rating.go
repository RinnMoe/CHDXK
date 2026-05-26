package handler

import (
	"context"

	"github.com/hibiken/asynq"

	"jcourse/internal/application"
)

type refreshCourseRatingScoresHandler struct {
	command *application.CourseRatingCommandService
}

func NewRefreshCourseRatingScoresHandler(command *application.CourseRatingCommandService) asynq.Handler {
	return &refreshCourseRatingScoresHandler{command: command}
}

func (h *refreshCourseRatingScoresHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	return h.command.RefreshRatingScores(ctx)
}
