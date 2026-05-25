package task

import (
	"context"
	"fmt"

	"jcourse/pkg/logx"
)

type asynqLogger struct{}

func newLogger() asynqLogger {
	return asynqLogger{}
}

func (l asynqLogger) Debug(args ...any) {
	logx.Debug(context.Background(), fmt.Sprint(args...))
}

func (l asynqLogger) Info(args ...any) {
	logx.Info(context.Background(), fmt.Sprint(args...))
}

func (l asynqLogger) Warn(args ...any) {
	logx.Warn(context.Background(), fmt.Sprint(args...))
}

func (l asynqLogger) Error(args ...any) {
	logx.Error(context.Background(), fmt.Sprint(args...))
}

func (l asynqLogger) Fatal(args ...any) {
	logx.Fatal(context.Background(), fmt.Sprint(args...))
}
