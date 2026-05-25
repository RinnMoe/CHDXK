package async

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hibiken/asynq"

	"jcourse/pkg/logx"
)

func TestTaskLoggingMiddlewareLogsSuccess(t *testing.T) {
	var buf bytes.Buffer
	logx.Configure(&buf)
	middleware := newTaskLoggingMiddleware()
	handler := middleware(asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		return nil
	}))

	if err := handler.ProcessTask(context.Background(), asynq.NewTask("test:success", nil)); err != nil {
		t.Fatalf("ProcessTask: %v", err)
	}

	entries := decodeLogEntries(t, &buf)
	if len(entries) != 2 {
		t.Fatalf("log entries = %d, want 2", len(entries))
	}
	if entries[0]["msg"] != "async task started" || entries[0]["type"] != "test:success" {
		t.Fatalf("start log = %#v", entries[0])
	}
	if entries[1]["msg"] != "async task completed" || entries[1]["type"] != "test:success" {
		t.Fatalf("complete log = %#v", entries[1])
	}
	if _, ok := entries[1]["duration"].(float64); !ok {
		t.Fatalf("duration = %#v, want milliseconds as number", entries[1]["duration"])
	}
}

func TestTaskLoggingMiddlewareLogsFailure(t *testing.T) {
	var buf bytes.Buffer
	logx.Configure(&buf)
	wantErr := errors.New("boom")
	middleware := newTaskLoggingMiddleware()
	handler := middleware(asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		return wantErr
	}))

	err := handler.ProcessTask(context.Background(), asynq.NewTask("test:failure", nil))
	if !errors.Is(err, wantErr) {
		t.Fatalf("ProcessTask error = %v, want %v", err, wantErr)
	}

	entries := decodeLogEntries(t, &buf)
	if len(entries) != 2 {
		t.Fatalf("log entries = %d, want 2", len(entries))
	}
	if entries[0]["msg"] != "async task started" {
		t.Fatalf("start log = %#v", entries[0])
	}
	if entries[1]["msg"] != "async task failed" || entries[1]["type"] != "test:failure" || entries[1]["err"] != "boom" {
		t.Fatalf("failure log = %#v", entries[1])
	}
	if _, ok := entries[1]["duration"].(float64); !ok {
		t.Fatalf("duration = %#v, want milliseconds as number", entries[1]["duration"])
	}
}

func decodeLogEntries(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var entries []map[string]any
	decoder := json.NewDecoder(buf)
	for decoder.More() {
		var entry map[string]any
		if err := decoder.Decode(&entry); err != nil {
			t.Fatalf("decode log entry: %v", err)
		}
		entries = append(entries, entry)
	}
	return entries
}
