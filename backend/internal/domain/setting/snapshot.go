package setting

import (
	"strconv"
	"strings"
	"time"
)

type Snapshot struct {
	values map[string]string
}

func NewSnapshot(values map[string]string) *Snapshot {
	items := make(map[string]string, len(values))
	for key, value := range values {
		items[key] = value
	}
	return &Snapshot{values: items}
}

func (s *Snapshot) String(key string) string {
	if s == nil {
		return ""
	}
	return s.values[key]
}

func (s *Snapshot) Int(key string) int {
	value, _ := strconv.Atoi(s.String(key))
	return value
}

func (s *Snapshot) Bool(key string) bool {
	value, _ := strconv.ParseBool(s.String(key))
	return value
}

func (s *Snapshot) Float(key string) float64 {
	value, _ := strconv.ParseFloat(s.String(key), 64)
	return value
}

func (s *Snapshot) Duration(key string) time.Duration {
	value, _ := time.ParseDuration(s.String(key))
	return value
}

func (s *Snapshot) StringList(key string) []string {
	parts := strings.FieldsFunc(s.String(key), func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t'
	})
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}
