package logx

import (
	"context"
	"io"
	"log/slog"
	"os"
)

func ConfigureDefault() {
	Configure(os.Stderr)
}

func Configure(w io.Writer) {
	if w == nil {
		w = os.Stderr
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(w, nil)))
}

func Logger() *slog.Logger {
	return slog.Default()
}

func Debug(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
}

func Fatal(ctx context.Context, msg string, args ...any) {
	Error(ctx, msg, args...)
	os.Exit(1)
}
