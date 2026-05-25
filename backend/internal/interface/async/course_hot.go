package async

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"jcourse/internal/application"
	"jcourse/internal/domain/course"
)

type recordHotCourseActivityHandler struct {
	command *application.CourseHotCommandService
}

func newRecordHotCourseActivityHandler(command *application.CourseHotCommandService) asynq.Handler {
	return &recordHotCourseActivityHandler{command: command}
}

func (h *recordHotCourseActivityHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload course.RecordHotCourseActivityPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	return h.command.RecordActivity(ctx, payload)
}
