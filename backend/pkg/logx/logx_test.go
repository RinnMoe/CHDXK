package logx

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"jcourse/pkg/requestid"
)

func TestInfoLogsRequestIDFromContext(t *testing.T) {
	var buf bytes.Buffer
	Configure(&buf)

	ctx := requestid.WithContext(context.Background(), "req-123")
	Info(ctx, "test message", "key", "value")

	var entry map[string]any
	if err := json.NewDecoder(&buf).Decode(&entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}
	if entry[requestid.GinKey] != "req-123" {
		t.Fatalf("request id = %#v", entry[requestid.GinKey])
	}
	if entry["key"] != "value" {
		t.Fatalf("key = %#v", entry["key"])
	}
}

func TestInfoDoesNotOverrideExplicitRequestID(t *testing.T) {
	var buf bytes.Buffer
	Configure(&buf)

	ctx := requestid.WithContext(context.Background(), "ctx-req")
	Info(ctx, "test message", requestid.GinKey, "explicit-req")

	var entry map[string]any
	if err := json.NewDecoder(&buf).Decode(&entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}
	if entry[requestid.GinKey] != "explicit-req" {
		t.Fatalf("request id = %#v", entry[requestid.GinKey])
	}
}
