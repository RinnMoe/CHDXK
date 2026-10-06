package teacher

import (
	"context"
	"strings"
	"time"
)

type Profile struct {
	URL        string            `json:"url"`
	Name       string            `json:"name"`
	Department string            `json:"department"`
	Title      string            `json:"title"`
	Fields     map[string]string `json:"fields"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type ProfileMatch struct {
	TeacherID   int       `json:"teacher_id"`
	Name        string    `json:"name"`
	Department  string    `json:"department"`
	Status      string    `json:"status"`
	SelectedURL string    `json:"selected_url"`
	Candidates  []Profile `json:"candidates"`
}

// Whitespace and equivalent parentheses do not establish a different identity.
// Do not fuzzy-match names or collapse different colleges.
func ProfileIdentity(name, department string) string {
	normalize := func(s string) string {
		return strings.Map(func(r rune) rune {
			if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\u3000' {
				return -1
			}
			if r == '（' {
				return '('
			}
			if r == '）' {
				return ')'
			}
			return r
		}, s)
	}
	return normalize(name) + "\x00" + normalize(department)
}

func ProfileCandidates(t TeacherView, teachers []TeacherView, profiles []Profile) ([]Profile, bool) {
	key := ProfileIdentity(t.Name, t.Department)
	localCount := 0
	for _, other := range teachers {
		if ProfileIdentity(other.Name, other.Department) == key {
			localCount++
		}
	}
	matches := []Profile{}
	seen := map[string]bool{}
	for _, p := range profiles {
		if ProfileIdentity(p.Name, p.Department) == key && !seen[p.URL] {
			matches = append(matches, p)
			seen[p.URL] = true
		}
	}
	return matches, localCount == 1 && len(matches) == 1
}

type ProfileSource interface {
	List(context.Context, func(int, int)) ([]Profile, error)
	Detail(context.Context, Profile) (Profile, error)
}
type ProfileStore interface {
	Teachers(context.Context) ([]TeacherView, error)
	Save(context.Context, []TeacherView, []Profile) error
	Matches(context.Context) ([]ProfileMatch, error)
	Bind(context.Context, int, string) error
	Edit(context.Context, int, Profile) error
	List(context.Context, TeacherFilter) ([]TeacherView, int64, error)
}
