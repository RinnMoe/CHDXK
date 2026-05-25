package async

import (
	"context"
	"time"

	"github.com/hibiken/asynq"

	"jcourse/pkg/logx"
)

func newTaskLoggingMiddleware() asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			startedAt := time.Now()
			logx.Info(ctx, "async task started", "type", t.Type())
			err := next.ProcessTask(ctx, t)
			duration := time.Since(startedAt)
			if err != nil {
				logx.Error(ctx, "async task failed", "type", t.Type(), "duration", duration, "err", err)
				return err
			}
			logx.Info(ctx, "async task completed", "type", t.Type(), "duration", duration)
			return nil
		})
	}
}
