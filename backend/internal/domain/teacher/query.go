package teacher

import "context"

type FilterItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type TeacherFilter struct {
	TeacherIDs []int
	Department string
	Title      string
	Q          string
	Code       string
	Name       string
	Pinyin     string
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

type TeacherFilters struct {
	Departments []FilterItem `json:"departments"`
	Titles      []FilterItem `json:"titles"`
}

// Read model interface
type TeacherQuery interface {
	FindBy(ctx context.Context, filter TeacherFilter) ([]TeacherView, int64, error)
	GetFilters(ctx context.Context) (*TeacherFilters, error)
}
