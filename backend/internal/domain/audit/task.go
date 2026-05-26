package audit

import (
	"context"
	"encoding/json"
	"time"

	"jcourse/internal/domain/task"
	"jcourse/pkg/logx"
)

const TaskTypeRecordLog = "audit:record_log"

type RecordLogPayload struct {
	OccurredAt  time.Time `json:"occurred_at"`
	ActorUserID int       `json:"actor_user_id"`
	Action      string    `json:"action"`
	TargetType  string    `json:"target_type"`
	TargetID    string    `json:"target_id"`
	Details     Details   `json:"details,omitempty"`
}

type RecordLogTask struct {
	payload RecordLogPayload
}

func NewRecordLogTask(log Log) RecordLogTask {
	return RecordLogTask{payload: RecordLogPayload{
		OccurredAt:  log.OccurredAt,
		ActorUserID: log.ActorUserID,
		Action:      log.Action,
		TargetType:  log.TargetType,
		TargetID:    log.TargetID,
		Details:     log.Details,
	}}
}

func (t RecordLogTask) Type() string {
	return TaskTypeRecordLog
}

func (t RecordLogTask) Payload() []byte {
	b, _ := json.Marshal(t.payload)
	return b
}

func EnqueueLog(ctx context.Context, entry Log) {
	if entry.OccurredAt.IsZero() {
		entry.OccurredAt = time.Now()
	}
	if entry.Details == nil {
		entry.Details = Details{}
	}
	if err := task.Enqueue(ctx, NewRecordLogTask(entry)); err != nil {
		logx.Warn(ctx, "enqueue audit log", "actor_user_id", entry.ActorUserID, "action", entry.Action, "target_type", entry.TargetType, "target_id", entry.TargetID, "err", err)
	}
}
