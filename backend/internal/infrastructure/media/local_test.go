package media

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func TestLocalStorageRoundTrip(t *testing.T) {
	root := t.TempDir()
	storage, err := NewLocalStorage(root)
	if err != nil {
		t.Fatalf("NewLocalStorage() = %v", err)
	}
	ctx := context.Background()
	key := "review/1/2026-10/sample.png"
	if err := storage.Put(ctx, key, bytes.NewReader([]byte("payload"))); err != nil {
		t.Fatalf("Put() = %v", err)
	}
	reader, err := storage.Open(ctx, key)
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	data, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil || string(data) != "payload" {
		t.Fatalf("read = %q err=%v", data, err)
	}
	if err := storage.Delete(ctx, key); err != nil {
		t.Fatalf("Delete() = %v", err)
	}
	if _, err := storage.Open(ctx, key); err == nil {
		t.Fatal("Open() after Delete should fail")
	}
}

func TestLocalStorageRejectsTraversal(t *testing.T) {
	storage, err := NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalStorage() = %v", err)
	}
	if err := storage.Put(context.Background(), "../escape.png", bytes.NewReader([]byte("x"))); err == nil {
		t.Fatal("Put() should reject traversal")
	}
}
