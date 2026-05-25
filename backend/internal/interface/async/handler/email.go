package handler

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	domainemail "jcourse/internal/domain/email"
)

type sendEmailHandler struct {
	service *domainemail.Service
}

func NewSendEmailHandler(service *domainemail.Service) asynq.Handler {
	return &sendEmailHandler{service: service}
}

func (h *sendEmailHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload domainemail.SendEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	return h.service.SendTemplatedEmail(ctx, payload)
}
