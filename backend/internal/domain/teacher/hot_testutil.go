//go:build test

package teacher

import "context"

type MockHotTeacherQuery struct {
	OnTopTeachers func(context.Context, string, string, int) ([]HotTeacherView, error)
}

func (m *MockHotTeacherQuery) TopTeachers(ctx context.Context, period, key string, limit int) ([]HotTeacherView, error) {
	return m.OnTopTeachers(ctx, period, key, limit)
}
