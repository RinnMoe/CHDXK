package handler

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"

	"jcourse/internal/domain/auth"
)

func TestFlushAccessHandlerCallsTracker(t *testing.T) {
	tracker := &fakeAsyncAccessTracker{}
	handler := NewFlushAccessHandler(tracker)

	if err := handler.ProcessTask(context.Background(), asynq.NewTask(auth.TaskTypeFlushAccess, auth.NewFlushAccessTask().Payload())); err != nil {
		t.Fatalf("ProcessTask: %v", err)
	}
	if !tracker.flushed {
		t.Fatal("expected tracker to be flushed")
	}
}

type fakeAsyncAccessTracker struct {
	flushed bool
}

func (t *fakeAsyncAccessTracker) RecordUserAccess(context.Context, int, time.Time) error {
	return nil
}

func (t *fakeAsyncAccessTracker) RecordApiKeyAccess(context.Context, int64, time.Time) error {
	return nil
}

func (t *fakeAsyncAccessTracker) Flush(context.Context) (auth.AccessFlushResult, error) {
	t.flushed = true
	return auth.AccessFlushResult{Users: 1, ApiKeys: 2}, nil
}
