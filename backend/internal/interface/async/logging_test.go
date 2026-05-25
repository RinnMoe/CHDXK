package async

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hibiken/asynq"
)

type fakeTaskLogger struct {
	lines []string
}

func (l *fakeTaskLogger) Printf(format string, v ...any) {
	l.lines = append(l.lines, strings.TrimSpace(fmt.Sprintf(format, v...)))
}

func TestTaskLoggingMiddlewareLogsSuccess(t *testing.T) {
	logger := &fakeTaskLogger{}
	middleware := newTaskLoggingMiddleware(logger)
	handler := middleware(asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		return nil
	}))

	if err := handler.ProcessTask(context.Background(), asynq.NewTask("test:success", nil)); err != nil {
		t.Fatalf("ProcessTask: %v", err)
	}

	if len(logger.lines) != 2 {
		t.Fatalf("log lines = %d, want 2", len(logger.lines))
	}
	if !strings.Contains(logger.lines[0], "async task started") || !strings.Contains(logger.lines[0], "type=test:success") {
		t.Fatalf("start log = %q", logger.lines[0])
	}
	if !strings.Contains(logger.lines[1], "async task completed") || !strings.Contains(logger.lines[1], "type=test:success") || !strings.Contains(logger.lines[1], "duration=") {
		t.Fatalf("complete log = %q", logger.lines[1])
	}
}

func TestTaskLoggingMiddlewareLogsFailure(t *testing.T) {
	logger := &fakeTaskLogger{}
	wantErr := errors.New("boom")
	middleware := newTaskLoggingMiddleware(logger)
	handler := middleware(asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		return wantErr
	}))

	err := handler.ProcessTask(context.Background(), asynq.NewTask("test:failure", nil))
	if !errors.Is(err, wantErr) {
		t.Fatalf("ProcessTask error = %v, want %v", err, wantErr)
	}

	if len(logger.lines) != 2 {
		t.Fatalf("log lines = %d, want 2", len(logger.lines))
	}
	if !strings.Contains(logger.lines[0], "async task started") {
		t.Fatalf("start log = %q", logger.lines[0])
	}
	if !strings.Contains(logger.lines[1], "async task failed") || !strings.Contains(logger.lines[1], "type=test:failure") || !strings.Contains(logger.lines[1], "error=boom") {
		t.Fatalf("failure log = %q", logger.lines[1])
	}
}
