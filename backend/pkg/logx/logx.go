package logx

import (
	"context"
	"io"
	"log/slog"
	"os"

	"jcourse/pkg/requestid"
)

func ConfigureDefault() {
	Configure(os.Stdout)
}

func Configure(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(w, nil)))
}

func Logger() *slog.Logger {
	return slog.Default()
}

func Debug(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, withRequestID(ctx, args)...)
}

func Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, withRequestID(ctx, args)...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, withRequestID(ctx, args)...)
}

func Error(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, withRequestID(ctx, args)...)
}

func Fatal(ctx context.Context, msg string, args ...any) {
	Error(ctx, msg, args...)
	os.Exit(1)
}

func withRequestID(ctx context.Context, args []any) []any {
	id, ok := requestid.FromContext(ctx)
	if !ok || hasRequestID(args) {
		return args
	}
	attrs := make([]any, 0, len(args)+2)
	attrs = append(attrs, args...)
	attrs = append(attrs, requestid.GinKey, id)
	return attrs
}

func hasRequestID(args []any) bool {
	for _, arg := range args {
		if key, ok := arg.(string); ok && key == requestid.GinKey {
			return true
		}
		if attr, ok := arg.(slog.Attr); ok && attr.Key == requestid.GinKey {
			return true
		}
	}
	return false
}
