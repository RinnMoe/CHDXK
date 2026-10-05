//go:build test

package course

import "context"

type MockCourseSource struct {
	ConnectFn func(context.Context) (SourceConnection, error)
}

func (m *MockCourseSource) Connect(ctx context.Context) (SourceConnection, error) {
	return m.ConnectFn(ctx)
}
