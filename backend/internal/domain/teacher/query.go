package teacher

import "context"

type TeacherFilter struct {
	Department string
	Title      string
	Pinyin     string // matches pinyin and pinyin_abbr with LIKE
	Page       int
	PageSize   int
}

type TeacherForQuery struct {
	ID         int
	Code       string
	Name       string
	Department string
	Title      string
	Pinyin     string
	PinyinAbbr string
}

type TeacherQuery interface {
	FindBy(ctx context.Context, filter TeacherFilter) ([]TeacherForQuery, int64, error)
}
