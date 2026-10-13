package course

import (
	"context"
	"encoding/json"
)

// SourceJobStore retains progress only. Authentication belongs to the temporary
// source connection and must never be stored here. Each owner has one latest job.
type SourceJobStore interface {
	Save(context.Context, int, bool, json.RawMessage) error
	Load(context.Context, int, bool) (json.RawMessage, error)
}
