package course

import (
	"context"
	"encoding/json"
	"time"
)

type SourceEvent struct {
	Type      string          `json:"type"`
	QR        string          `json:"qr,omitempty"`
	Semester  string          `json:"semester,omitempty"`
	Semesters []string        `json:"semesters,omitempty"`
	Error     string          `json:"error,omitempty"`
	Document  json.RawMessage `json:"document,omitempty"`
	Report    json.RawMessage `json:"report,omitempty"`
	Progress  *SourceProgress `json:"progress,omitempty"`
}

// SourceProgress contains counts only, never upstream URLs or credentials.
type SourceProgress struct {
	Stage     string    `json:"stage"`
	Completed int       `json:"completed"`
	Total     *int      `json:"total,omitempty"`
	Failed    int       `json:"failed"`
	Retries   int       `json:"retries"`
	Page      int       `json:"page,omitempty"`
	Pages     int       `json:"pages,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SourceConnection holds one temporary, authenticated HTTP collector session.
type SourceConnection interface {
	Events() <-chan SourceEvent
	Sync([]string) error
	Close()
}

type CourseSource interface {
	Connect(context.Context) (SourceConnection, error)
}
