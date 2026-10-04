package teacher

import "context"

// Teacher heat is a read model of existing course activity, not another score store.
type HotTeacherView struct {
	Teacher     TeacherView
	Score       int64
	CourseCount int
}

type HotTeacherQuery interface {
	TopTeachers(ctx context.Context, period, periodKey string, limit int) ([]HotTeacherView, error)
}
