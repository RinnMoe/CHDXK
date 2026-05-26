package handler

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"jcourse/internal/application"
	"jcourse/internal/domain/audit"
)

type recordAuditLogHandler struct {
	command *application.AuditLogCommandService
}

func NewRecordAuditLogHandler(command *application.AuditLogCommandService) asynq.Handler {
	return &recordAuditLogHandler{command: command}
}

func (h *recordAuditLogHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload audit.RecordLogPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	return h.command.Record(ctx, payload)
}
