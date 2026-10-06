//go:build test

package teacher

import "context"

type MockProfileSource struct {
	OnList   func(context.Context, func(int, int)) ([]Profile, error)
	OnDetail func(context.Context, Profile) (Profile, error)
	OnFind   func(context.Context, string, string) ([]Profile, error)
}

func (m *MockProfileSource) List(ctx context.Context, progress func(int, int)) ([]Profile, error) {
	return m.OnList(ctx, progress)
}
func (m *MockProfileSource) Detail(ctx context.Context, p Profile) (Profile, error) {
	return m.OnDetail(ctx, p)
}

type MockProfileStore struct {
	OnTeachers func(context.Context) ([]TeacherView, error)
	OnSave     func(context.Context, []TeacherView, []Profile) error
	OnMatches  func(context.Context) ([]ProfileMatch, error)
	OnBind     func(context.Context, int, string) error
	OnEdit     func(context.Context, int, Profile) error
	OnList     func(context.Context, TeacherFilter) ([]TeacherView, int64, error)
}

func (m *MockProfileStore) Teachers(ctx context.Context) ([]TeacherView, error) {
	return m.OnTeachers(ctx)
}
func (m *MockProfileStore) Save(ctx context.Context, t []TeacherView, p []Profile) error {
	return m.OnSave(ctx, t, p)
}
func (m *MockProfileStore) Matches(ctx context.Context) ([]ProfileMatch, error) {
	return m.OnMatches(ctx)
}
func (m *MockProfileStore) Bind(ctx context.Context, id int, url string) error {
	return m.OnBind(ctx, id, url)
}

func (m *MockProfileStore) Edit(ctx context.Context, id int, p Profile) error {
	return m.OnEdit(ctx, id, p)
}

func (m *MockProfileStore) List(ctx context.Context, f TeacherFilter) ([]TeacherView, int64, error) {
	if m.OnList != nil {
		return m.OnList(ctx, f)
	}
	return nil, 0, nil
}

func (m *MockProfileSource) Find(ctx context.Context, name, department string) ([]Profile, error) {
	return m.OnFind(ctx, name, department)
}
