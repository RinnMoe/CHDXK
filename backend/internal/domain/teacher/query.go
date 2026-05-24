package teacher

import (
	"context"
	"strings"

	pinyin "github.com/mozillazg/go-pinyin"
)

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
}

func NewSearchName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	normalArgs := pinyin.NewArgs()
	normalArgs.Style = pinyin.Normal
	normal := pinyin.LazyPinyin(name, normalArgs)
	spaced := strings.Join(normal, " ")
	compact := strings.Join(normal, "")

	abbrArgs := pinyin.NewArgs()
	abbrArgs.Style = pinyin.FirstLetter
	abbr := strings.Join(pinyin.LazyPinyin(name, abbrArgs), "")

	parts := make([]string, 0, 4)
	seen := make(map[string]struct{}, 4)
	for _, part := range []string{name, spaced, compact, abbr} {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		parts = append(parts, part)
	}
	return strings.Join(parts, " ")
}

type TeacherFilters struct {
	Departments []FilterItem `json:"departments"`
	Titles      []FilterItem `json:"titles"`
}

// Read model interface
type TeacherQuery interface {
	FindBy(ctx context.Context, filter TeacherFilter) ([]TeacherView, int64, error)
	GetByID(ctx context.Context, teacherID int) (*TeacherView, error)
	GetFilters(ctx context.Context) (*TeacherFilters, error)
}
