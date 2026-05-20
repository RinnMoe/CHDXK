package teacher

import "context"

type TeacherFilter struct {
	Department string
	Title      string
	Pinyin     string // matches pinyin and pinyin_abbr with LIKE
	Page       int
	PageSize   int
}

// Read model: teacher query result
type TeacherView struct {
	ID         int
	Code       string
	Name       string
	Department string
	Title      string
	Pinyin     string
	PinyinAbbr string
}

// Read model interface
type TeacherQuery interface {
	FindBy(ctx context.Context, filter TeacherFilter) ([]TeacherView, int64, error)
}
