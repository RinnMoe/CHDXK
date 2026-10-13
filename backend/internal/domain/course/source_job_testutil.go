//go:build test

package course

import (
	"context"
	"encoding/json"
)

type MockSourceJobStore struct {
	SaveFn func(context.Context, int, bool, json.RawMessage) error
	LoadFn func(context.Context, int, bool) (json.RawMessage, error)
}

func (m *MockSourceJobStore) Save(ctx context.Context, owner int, personal bool, value json.RawMessage) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, owner, personal, value)
	}
	return nil
}
func (m *MockSourceJobStore) Load(ctx context.Context, owner int, personal bool) (json.RawMessage, error) {
	if m.LoadFn != nil {
		return m.LoadFn(ctx, owner, personal)
	}
	return nil, nil
}
