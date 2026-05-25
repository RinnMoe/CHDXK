package async

import (
	"context"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

type taskLogger interface {
	Printf(format string, v ...any)
}

func newTaskLoggingMiddleware(logger taskLogger) asynq.MiddlewareFunc {
	if logger == nil {
		logger = log.Default()
	}
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			startedAt := time.Now()
			logger.Printf("async task started: type=%s", t.Type())
			err := next.ProcessTask(ctx, t)
			duration := time.Since(startedAt)
			if err != nil {
				logger.Printf("async task failed: type=%s duration=%s error=%v", t.Type(), duration, err)
				return err
			}
			logger.Printf("async task completed: type=%s duration=%s", t.Type(), duration)
			return nil
		})
	}
}
