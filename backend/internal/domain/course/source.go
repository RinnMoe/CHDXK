package course

import (
	"context"
	"encoding/json"
)

type SourceEvent struct {
	Type      string          `json:"type"`
	QR        string          `json:"qr,omitempty"`
	Semester  string          `json:"semester,omitempty"`
	Semesters []string        `json:"semesters,omitempty"`
	Error     string          `json:"error,omitempty"`
	Document  json.RawMessage `json:"document,omitempty"`
	Report    json.RawMessage `json:"report,omitempty"`
}

// SourceConnection holds one temporary, authenticated collector process.
type SourceConnection interface {
	Events() <-chan SourceEvent
	Sync([]string) error
	Close()
}

type CourseSource interface {
	Connect(context.Context) (SourceConnection, error)
}
