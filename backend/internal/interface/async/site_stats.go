package async

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"jcourse/internal/application"
	"jcourse/internal/domain/stat"
)

type collectDailySiteStatsHandler struct {
	command *application.SiteStatsCommandService
}

func newCollectDailySiteStatsHandler(command *application.SiteStatsCommandService) asynq.Handler {
	return &collectDailySiteStatsHandler{command: command}
}

func (h *collectDailySiteStatsHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload stat.CollectDailySiteStatsPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	_, err := h.command.CollectDailyByDateString(ctx, payload.StatDate)
	return err
}
