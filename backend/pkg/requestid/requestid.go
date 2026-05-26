package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

const (
	HeaderRequestID = "X-Request-Id"
	GinKey          = "request_id"
)

type contextKey struct{}

func New() string {
	prefix := time.Now().UTC().Format("20060102150405")

	var b [10]byte
	if _, err := rand.Read(b[:]); err != nil {
		return prefix
	}
	return prefix + hex.EncodeToString(b[:])
}

func WithContext(ctx context.Context, id string) context.Context {
	id = strings.TrimSpace(id)
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, contextKey{}, id)
}

func FromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	id, ok := ctx.Value(contextKey{}).(string)
	return id, ok && id != ""
}
