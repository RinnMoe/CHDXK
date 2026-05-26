package handler

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	domainemail "jcourse/internal/domain/email"
)

type sendEmailHandler struct {
	sender domainemail.Sender
}

func NewSendEmailHandler(sender domainemail.Sender) asynq.Handler {
	return &sendEmailHandler{sender: sender}
}

func (h *sendEmailHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload domainemail.SendEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	return h.sender.SendEmail(ctx, payload.Email)
}
