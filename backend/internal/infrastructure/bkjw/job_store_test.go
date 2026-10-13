//go:build test

package bkjw

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestJobStoreIsolatesOwnersAndCollectorKinds(t *testing.T) {
	root := t.TempDir()
	store := NewJobStore(root)
	value := json.RawMessage(`{"id":"test","state":"running"}`)
	if err := store.Save(context.Background(), 1, true, value); err != nil {
		t.Fatal(err)
	}
	restarted := NewJobStore(root)
	got, err := restarted.Load(context.Background(), 1, true)
	if err != nil || string(got) != string(value) {
		t.Fatal(string(got), err)
	}
	for _, key := range []struct {
		owner    int
		personal bool
	}{{2, true}, {1, false}} {
		got, err := restarted.Load(context.Background(), key.owner, key.personal)
		if err != nil || got != nil {
			t.Fatal("cross-owner/mode data", got, err)
		}
	}
	if err := store.Save(context.Background(), 1, true, json.RawMessage(`broken`)); err == nil {
		t.Fatal("invalid JSON saved")
	}
}

func TestOldConnectionCannotOverwriteReplacementJob(t *testing.T) {
	store := NewJobStore(t.TempDir())
	now := time.Now()
	newer, _ := json.Marshal(map[string]any{"id": "new-job", "created_at": now})
	older, _ := json.Marshal(map[string]any{"id": "old-job", "created_at": now.Add(-time.Minute)})
	if err := store.Save(context.Background(), 1, true, newer); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), 1, true, older); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load(context.Background(), 1, true)
	if err != nil || string(got) != string(newer) {
		t.Fatal("stale job replaced current job", string(got), err)
	}
}
